package main

import (
	"errors"
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

type TransactionData struct {
	FirstName       string
	LastName        string
	Email           string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour        string
	ExpiryMonth     int
	ExpiryYear      int
	BankReturnCode  string
}

// gets the transac data from stripe and post
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return txnData, err

	}

	firstName := r.Form.Get("first_name")
	lastName := r.Form.Get("last_name")
	email := r.Form.Get("cardholder_email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")

	if paymentAmount == "" {
		return txnData, errors.New("payment_amount missing")
	}

	paymentCurrency := r.Form.Get("payment_currency")

	amount, _ := strconv.Atoi(paymentAmount)

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	pi, err := card.GetPaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)

	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4

	expiryMonth := pm.Card.ExpMonth

	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:        lastFour,
		ExpiryMonth:     int(expiryMonth),
		ExpiryYear:      int(expiryYear),
		BankReturnCode:  pi.Charges.Data[0].ID,
	}

	return txnData, nil

}

func (app *application) PaymentReceipt(w http.ResponseWriter, r *http.Request) {

	app.infoLog.Printf("raw payment_amount = %q", r.Form.Get("payment_amount"))

	err := r.ParseForm()

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	teaID, _ := strconv.Atoi(r.Form.Get("product_id"))

	txnData, err := app.GetTransactionData(r)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	//create new customer

	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	app.infoLog.Print(customerID)

	//finally create new transaction

	txn := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:         int(txnData.ExpiryMonth),
		ExpiryYear:          int(txnData.ExpiryYear),
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
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
		Amount:        txnData.PaymentAmount,
		TimeCreated:   time.Now(),
		TimeUpdated:   time.Now(),
	}

	_, err = app.SaveOrder(order)

	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// =========== ============

	// redirection

	app.Session.Put(r.Context(), "receipt", txnData)

	http.Redirect(w, r, "/receipt", http.StatusSeeOther)

}

func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {
	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")
	if err := app.renderTemplate(w, r, "receipt", &templateData{
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
