package main

import (
	"log"
	"net/http"

	"github.com/caddyserver/certmagic"
	"github.com/kelseyhightower/envconfig"
)

var db *DB
var cfg Config

func main() {
	var err error
	err = envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err = NewDB()
	if err != nil {
		log.Fatal("Error initializing DB: ", err)
	}

	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.Email = cfg.Email

	mux := http.NewServeMux()

	mux.HandleFunc("/user", apiKeyRequired(handleUser))
	mux.HandleFunc("/webhook", handleWebhook)

	log.Fatal(certmagic.HTTPS([]string{cfg.Domain}, mux))
}
