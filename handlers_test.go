package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var testUser user

func deleteTestUser() {
	db.db.Exec("DELETE FROM users WHERE id = ?", "test")
}

func TestAPIKey(t *testing.T) {

	handler := apiKeyRequired(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Test missing api key", func(t *testing.T) {
		r := httptest.NewRequest(
			http.MethodPost,
			"/user?id=test",
			nil,
		)
		w := httptest.NewRecorder()
		handler(w, r)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusUnauthorized {
			t.Fatal("Expected", http.StatusUnauthorized, " got", res.StatusCode)
		}
	})

	t.Run("Test has api key", func(t *testing.T) {
		r := httptest.NewRequest(
			http.MethodPost,
			"/user?id=test",
			nil,
		)
		r.Header.Add("api-key", cfg.ApiKey)
		w := httptest.NewRecorder()
		handler(w, r)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatal("Expected", http.StatusOK, " got", res.StatusCode)
		}
	})
}

func TestUserHandler(t *testing.T) {
	t.Run("Test paid", func(t *testing.T) {
		created, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
		paidUser := user{
			id:      "test",
			paid:    true,
			created: created,
		}
		if err := db.InsertUser(paidUser); err != nil {
			panic("User insert failed")
		}
		defer deleteTestUser()

		r := httptest.NewRequest(
			http.MethodPost,
			"/user?id=test",
			nil,
		)
		w := httptest.NewRecorder()
		handleUser(w, r)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatal("Request failed with", res.StatusCode)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("expected error to be nil got %v", err)
		}
		j := make(map[string]interface{})
		err = json.Unmarshal(data, &j)
		if err != nil {
			t.Fatal("Unmarshal failed", err)
		}
		expected := map[string]interface{}{"status": "paid"}
		if !cmp.Equal(expected, j) {
			t.Fatal("Expected", expected, " got", j)
		}
	})

	t.Run("Test trial", func(t *testing.T) {
		created := time.Now()
		trialUser := user{
			id:      "test",
			paid:    false,
			created: created,
		}
		if err := db.InsertUser(trialUser); err != nil {
			panic("User insert failed")
		}
		defer deleteTestUser()

		r := httptest.NewRequest(
			http.MethodPost,
			"/user?id=test",
			nil,
		)
		w := httptest.NewRecorder()
		handleUser(w, r)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatal("Request failed with", res.StatusCode)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("expected error to be nil got %v", err)
		}
		j := make(map[string]interface{})
		err = json.Unmarshal(data, &j)
		if err != nil {
			t.Fatal("Unmarshal failed", err)
		}
		expected := map[string]interface{}{"status": "trial"}
		if !cmp.Equal(expected, j) {
			t.Fatal("Expected", expected, " got", j)
		}
	})

	t.Run("Test trial ended", func(t *testing.T) {
		created := time.Now().AddDate(0, 0, int(-cfg.TrialDays-1))
		trialUser := user{
			id:      "test",
			paid:    false,
			created: created,
		}
		if err := db.InsertUser(trialUser); err != nil {
			panic("User insert failed")
		}
		defer deleteTestUser()

		r := httptest.NewRequest(
			http.MethodPost,
			"/user?id=test",
			nil,
		)
		w := httptest.NewRecorder()
		handleUser(w, r)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatal("Request failed with", res.StatusCode)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("expected error to be nil got %v", err)
		}
		j := make(map[string]interface{})
		err = json.Unmarshal(data, &j)
		if err != nil {
			t.Fatal("Unmarshal failed", err)
		}
		expected := map[string]interface{}{"status": "trial_ended", "payment_link": cfg.PaymentLink}
		if !cmp.Equal(expected, j) {
			t.Fatal("Expected", expected, " got", j)
		}
	})
}
