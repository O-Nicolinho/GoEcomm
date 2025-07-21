package models

import (
	"context"
	"database/sql"
	"time"
)

// this is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// this works as a wrapper for our models
type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

type Teas struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	InventoryAmt int       `json:"inventory_amount"`
	TimeCreated  time.Time `json:"-"`
	Image        string    `json:"image"`
	TimeUpdated  time.Time `json:"-"`
	Price        int       `json:"price"`
}

// type for all orders
type Order struct {
	ID            int       `json:"id"`
	TeaID         int       `json:"tea_id"`
	TransactionID int       `json:"transaction_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	TimeCreated   time.Time `json:"-"`
	TimeUpdated   time.Time `json:"-"`
}

type Status struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	TimeCreated time.Time `json:"-"`
	TimeUpdated time.Time `json:"-"`
}

type TransactionStatus struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	TimeCreated time.Time `json:"-"`
	TimeUpdated time.Time `json:"-"`
}

type Transaction struct {
	ID                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusID int       `json:"transaction_status_id"`
	TimeCreated         time.Time `json:"-"`
	TimeUpdated         time.Time `json:"-"`
}

type User struct {
	ID          int       `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	TimeCreated time.Time `json:"-"`
	TimeUpdated time.Time `json:"-"`
}

func (m *DBModel) GetTea(id int) (Teas, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tea Teas

	row := m.DB.QueryRowContext(ctx, "select id, name from tea_stock where id = ?", id)
	err := row.Scan(&tea.ID, &tea.Name)
	if err != nil {
		return tea, err
	}

	return tea, nil
}
