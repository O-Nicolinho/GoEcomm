package main

import (
	"net/http"
	"strconv"

	"github.com/O-Nicolinho/GoEcomm/internal/cards"
	"github.com/go-chi/chi/v5"
)

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "terminal", &templateData{
		StringMap: stringMap,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) PaymentReceipt(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	cardHolder := r.Form.Get("cardholder_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.GetPaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4

	expiryMonth := pm.Card.ExpMonth

	expiryYear := pm.Card.ExpYear

	data := make(map[string]interface{})

	data["cardholder"] = cardHolder
	data["email"] = email
	data["pi"] = paymentIntent
	data["pm"] = paymentMethod
	data["pa"] = paymentAmount
	data["pc"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	if err := app.renderTemplate(w, r, "succeeded", &templateData{
		Data: data,
	}); err != nil {
		app.errorLog.Println(err)
	}

}

// this func displays the page to buy some tea
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	teaID, _ := strconv.Atoi(id)
	tea, err := app.DB.GetTea(teaID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["tea"] = tea

	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}
