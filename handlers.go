package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

func apiKeyRequired(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != cfg.ApiKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		f(w, r)
	}
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var vals = r.URL.Query()
	var id = vals.Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u, err := db.GetUser(id)
	if errors.Is(err, ErrNotFound) {
		u = user{id: id, created: time.Now(), paid: false}
		err = db.InsertUser(u)
	}

	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(map[string]string)
	if u.paid {
		resp["status"] = "paid"
	} else if u.created.Before(time.Now().AddDate(0, 0, int(-cfg.TrialDays))) {
		resp["status"] = "trial_ended"
		resp["payment_link"] = cfg.PaymentLink
	} else {
		resp["status"] = "trial"
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsonResp)
}

// https://stripe.com/docs/webhooks
func handleWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// If you are testing your webhook locally with the Stripe CLI you
	// can find the endpoint's secret by running `stripe listen`
	// Otherwise, find your endpoint's secret in your webhook settings
	// in the Developer Dashboard
	endpointSecret := cfg.EndpointSecret

	// Pass the request body and Stripe-Signature header to ConstructEvent, along
	// with the webhook signing key.
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"),
		endpointSecret)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "checkout.session.completed":
		// https://stripe.com/docs/api/events/types#event_types-checkout.session.completed
		// this will contain the client_reference_id i.e. the unique device ID from garmin
		var cs stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &cs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if cs.ClientReferenceID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		success := db.SetPaid(cs.ClientReferenceID)
		if !success {
			fmt.Fprintf(os.Stderr, "Checkout session failed for: %s\n", cs.ClientReferenceID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(os.Stdout, "Checkout session completed for: %s\n", cs.ClientReferenceID)
	default:
		// unhandled
	}

	w.WriteHeader(http.StatusOK)
}
