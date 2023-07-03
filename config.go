package main

import (
	"errors"
	"strconv"
)

type Config struct {
	ApiKey         string    `required:"true" envconfig: "API_KEY`          // Generated with `go run scripts/api-key.go`
	Email          string    `required:"true" envconfig: "EMAIL"`           // Email for certificate
	Domain         string    `required:"true" envconfig: "DOMAIN"`          // Domain for certificate
	TrialDays      TrialDays `default:7 envconfig: "TRIAL_DAYS"`            // Length of trial period
	EndpointSecret string    `required:"true" envconfig: "ENDPOINT_SECRET"` // Stripe endpoint secret for webhooks
	PaymentLink    string    `required:"true" envconfig: "PAYMENT_LINK"`    // Stripe payment link
	DBPath         string    `required:"true" envconfig: DB_PATH"`          // Path to sqlite db file
}

type TrialDays int

func (td TrialDays) Decode(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if v < 0 {
		return errors.New("TRIAL_DAYS env var must be >= 0")
	}
	td = TrialDays(v)
	return nil
}
