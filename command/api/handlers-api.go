package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/O-Nicolinho/GoEcomm/internal/cards"
)

// info we're receiving from frontend
type stripePayload struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	ID      int    `json:"id,omitempty"`
}

func (app *application) GetPaymentIntent(w http.ResponseWriter, r *http.Request) {
	// 1. Decode JSON coming from the browser
	var p stripePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		app.errorLog.Println("decode:", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 2. Convert amount (string → int cents)
	amountInt, err := strconv.Atoi(p.Amount)
	if err != nil {
		http.Error(w, "amount must be a number", http.StatusBadRequest)
		return
	}

	// 3. Create / charge the intent
	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: p.Currency,
	}

	intent, msg, err := card.Charge(p.Currency, amountInt)

	// Always JSON:
	w.Header().Set("Content-Type", "application/json")

	if err == nil && intent != nil {
		// ----- SUCCESS (HTTP 200) -----
		json.NewEncoder(w).Encode(struct {
			OK           bool   `json:"ok"`
			ClientSecret string `json:"client_secret"`
		}{
			OK: true, ClientSecret: intent.ClientSecret,
		})
		return
	}

	// ----- ERROR (HTTP 400) -----
	app.errorLog.Println("stripe:", err)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(struct {
		OK      bool   `json:"ok"`
		Message string `json:"message"`
	}{
		OK: false, Message: msg, // e.g. “Your card was declined”
	})
}
