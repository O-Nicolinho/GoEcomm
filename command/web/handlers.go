package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/O-Nicolinho/GoEcomm/internal/cards"
	"github.com/O-Nicolinho/GoEcomm/internal/models"
	"github.com/go-chi/chi/v5"
)

// displays the vt
func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "terminal", &templateData{
		StringMap: stringMap,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
	}
}

// displays the home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	stringMap := make(map[string]string)
	stringMap["publishable_key"] = app.config.stripe.key

	if err := app.renderTemplate(w, r, "home", &templateData{
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

	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")
	teaID, _ := strconv.Atoi(r.Form.Get("product_id"))

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

	//create new customer

	customerID, err := app.SaveCustomer(firstName, lastName, email)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Print(customerID)

	//finally create new transaction

	amount, err := strconv.Atoi(paymentAmount)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	txn := models.Transaction{
		Amount:              amount,
		Currency:            paymentCurrency,
		LastFour:            lastFour,
		ExpiryMonth:         int(expiryMonth),
		ExpiryYear:          int(expiryYear),
		BankReturnCode:      pi.Charges.Data[0].ID,
		TransactionStatusID: 2,
	}

	txnID, err := app.SaveTransaction(txn)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create new order

	order := models.Order{
		TeaID:         teaID,
		TransactionID: txnID,
		CustomerID:    customerID,
		StatusID:      1,
		Quantity:      1,
		Amount:        amount,
		TimeCreated:   time.Now(),
		TimeUpdated:   time.Now(),
	}

	_, err = app.SaveOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// =========== ============

	data := make(map[string]interface{})

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

func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (app *application) SaveTransaction(txn models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction((txn))

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)

	if err != nil {
		return 0, err
	}
	return id, nil
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
