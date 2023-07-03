package main

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	cfg = Config{
		ApiKey:         "test",
		Email:          "test@test.com",
		Domain:         "test.com",
		TrialDays:      7,
		EndpointSecret: "secret",
		PaymentLink:    "https://buy.stripe.com/test",
		DBPath:         "./db/test.db",
	}

	var err error
	db, err = NewDB()
	if err != nil {
		log.Fatal("Failed to init db:", err)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}

func teardown() {
	db.db.Exec("DELETE FROM users")
}

func TestDB(t *testing.T) {
	created, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
	user := user{
		id:      "test",
		paid:    false,
		created: created,
	}

	t.Run("Insert new user", func(t *testing.T) {
		defer teardown()
		err := db.InsertUser(user)
		if err != nil {
			t.Fatal("Insert failed")
		}
		u, err := db.GetUser(user.id)
		if err != nil {
			t.Fatal("Select failed:", err)
		}
		if !cmp.Equal(u, user, cmp.AllowUnexported(user)) {
			t.Fatal("Expected user not same as actual user")
		}
	})

	t.Run("Insert user exists", func(t *testing.T) {
		defer teardown()
		update := user
		err := db.InsertUser(update)
		if err != nil {
			t.Fatal("Insert failed")
		}
		u, err := db.GetUser(update.id)
		if err != nil {
			t.Fatal("Select failed:", err)
		}
		if !cmp.Equal(u, update, cmp.AllowUnexported(update)) {
			t.Fatal("Expected user not same as actual user")
		}
	})

	t.Run("User paid", func(t *testing.T) {
		defer teardown()
		if err := db.InsertUser(user); err != nil {
			t.Fatal("Insert failed")
		}
		res := db.SetPaid(user.id)
		if res != true {
			t.Fatal("Paid failed")
		}
		u, err := db.GetUser(user.id)
		if err != nil {
			t.Fatal("Select failed:", err)
		}
		if u.paid != true {
			t.Fatal("Expected paid to be true")
		}
	})
}
