package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db             *sql.DB
	insertUserStmt *sql.Stmt
	updatePaidStmt *sql.Stmt
	selectUserStmt *sql.Stmt
}

const create string = `
  CREATE TABLE IF NOT EXISTS 
		users (
			id TEXT NOT NULL PRIMARY KEY,
			created DATETIME NOT NULL,
			paid INTEGER NOT NULL DEFAULT 0
		)
	;`

func NewDB() (*DB, error) {
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create); err != nil {
		return nil, err
	}

	insertUserStmt, err := db.Prepare("INSERT INTO users (id, created, paid) VALUES(?,?,?)")
	if err != nil {
		log.Fatal("Error preparing insert user statement:", err)
	}

	updatePaidStmt, err := db.Prepare("UPDATE users SET paid = 1 WHERE id = ?")
	if err != nil {
		log.Fatal("Error preparing updatePaid statement:", err)
	}

	selectUserStmt, err := db.Prepare("SELECT id, created, paid FROM users WHERE id=?")
	if err != nil {
		log.Fatal("Error preparing selectUser statement:", err)
	}

	_, err = db.Exec(`
		pragma journal_mode = WAL;
		pragma synchronous = normal;
	`)
	if err != nil {
		log.Fatal("Error running db performance statements:", err)
	}

	return &DB{
		db:             db,
		insertUserStmt: insertUserStmt,
		updatePaidStmt: updatePaidStmt,
		selectUserStmt: selectUserStmt,
	}, nil
}

type user struct {
	id      string
	created time.Time
	paid    bool
}

type UserExistsError struct {
	user user
}

func (e *UserExistsError) Error() string {
	return fmt.Sprint("User already exists:", e.user)
}

type DBWriteError struct {
	cause error
}

func (e *DBWriteError) Error() string {
	return fmt.Sprint("DB write error: ", e.cause)
}

func (c *DB) InsertUser(u user) error {
	_, err := c.insertUserStmt.Exec(
		u.id,
		u.created.Format("2006-01-02T15:04:05"),
		u.paid,
	)
	if err != nil {
		sqliteErr := err.(sqlite3.Error)
		if sqliteErr.Code == sqlite3.ErrConstraint {
			log.Println("User update failed:", err)
			return &UserExistsError{u}
		} else {
			log.Println("User insert failed:", err)
			return &DBWriteError{err}
		}
	}
	return nil
}

func randomSleep(min int, max int) {
	time.Sleep(time.Duration(rand.Intn(max-min+1) + min))
}

func (c *DB) SetPaid(id string) bool {
	_, err := c.updatePaidStmt.Exec(id)
	attempts := 1
	for err != nil && attempts <= 20 {
		log.Println("Update payment status failed. user:", id, ", attempts:", attempts, " error:", err)
		_, err = c.updatePaidStmt.Exec(id)
		attempts += 1
		// random sleep between 200ms and 1 sec to decorrelate retries
		randomSleep(200, 1000)
	}
	if err != nil && attempts >= 20 {
		log.Println("Update payment status: all attempts failed. user:", id, " error:", err)
		return false
	}
	return true
}

var ErrNotFound = errors.New("user not found")

func (c *DB) GetUser(id string) (user, error) {
	var user user
	err := c.selectUserStmt.QueryRow(id).Scan(&user.id, &user.created, &user.paid)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrNotFound
		}
		log.Println("User select failed:", err)
		return user, err
	}
	return user, nil
}
