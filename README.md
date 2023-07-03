# Overview

Provides an implementation to integrate a Garmin device app with stripe for payments processing. Users are identified using a unique ID generated on the Garmin device for the application. Each time a user opens the application on the watch, a web request can be made to this webserver to determine their payment status: trial, trial_ended or paid. After the configurable trial period has ended, payments are requested using a Stripe payment link which can be opened on the Garmin device in a Browser. The webserver is informed of a successful payment via a Stripe webhook and marks a user as paid in the SQLite database. 


# Generate a secure API key
go run scripts/api-key.go
