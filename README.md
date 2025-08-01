# LT Tea Company Web App

## =============== Features ===============

- **Secure Stripe Checkout**: payments and instant receipts are processed through Stripe’s Payment Intents API, with client-side card elements and server-side intent confirmation.

- **Inventory and Order Integration**: each purchase runs inside a single SQL transaction that firstly inserts the order, then decrements stock, and finally rolls back automatically if inventory is insufficient.

- **JWT-style Auth Tokens**: time-limited, SHA 256-hashed tokens (generated in models/tokens.go) login and logout without storing sessions in memory.

- **Bootstrap 5 UI**: implements the frontend with Go's html/template, and uses a custom CSS to make the design of the frontend consistent with the brand's colour scheme.

- **Simple Static Asset Serving**: static assets served with Go’s http.FileServer so the app remains a single binary + static/ folder.

- **Types for DB Interfacing**: All SQL commands exist in the models.go file, making the codebase more modular and organized.

- **Graceful Restarts**: A new context.Context is passed to every request, ensuring that whenever a shutdown starts, new requests are denied but requests that are in progress are allowed to finish.

- **Pluggable SMTP Mailer**: contact-form messages are relayed via any SMTP host (SendGrid, MailHog for dev, etc.)

- **12-factor config**: so Stripe, SMTP, and DB secrets injected via environment variables which ensures that no credentials are committed to the repo.


## =============== Demo ===============

![demo](docs/)

