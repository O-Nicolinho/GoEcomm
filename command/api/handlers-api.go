package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/O-Nicolinho/LT-TeaCompany-WebApp/internal/cards"
	"github.com/O-Nicolinho/LT-TeaCompany-WebApp/internal/models"
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

func (app *application) GetTeaByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	teaID, _ := strconv.Atoi(id)
	tea, err := app.DB.GetTea(teaID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	out, err := json.MarshalIndent(tea, "", "  ")

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &userInput)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// get usr from DB by email

	user, err := app.DB.GetUserByEmail(userInput.Email)

	if err != nil {
		app.invalidCredentials(w)
		return
	}

	// validate the input password, send an err if it's invalid

	validPassword, err := app.passwordMatches(user.Password, userInput.Password)

	if err != nil {
		app.invalidCredentials(w)
		return
	}

	if !validPassword {
		app.invalidCredentials(w)
		return
	}

	// now, if we have a valid user email & the password is also valid for
	// that user, we can generate the token

	token, err := models.GenerateToken(user.ID, 24*time.Hour, models.ScopeAuthentication)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// now we save it to the db

	err = app.DB.InsertToken(token, user)

	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	var payload struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Token   *models.Token `json:"authentication_token"`
	}

	payload.Error = false

	payload.Message = fmt.Sprintf("token for %s created.", userInput.Email)

	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)
}
