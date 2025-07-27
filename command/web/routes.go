package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(SessionLoad)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Get("/", app.Home)

	mux.Get("/donate", app.VirtualTerminal)
	mux.Post("/payment-succeeded", app.PaymentReceipt)
	mux.Get("/receipt", app.Receipt)
	mux.Get("/shop", app.Shop)

	mux.Get("/tea/{id}", app.ChargeOnce)

	// auth routes

	mux.Get("/login", app.LoginPage)

	return mux
}
