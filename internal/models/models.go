package models

import (
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
	TimeUpdated  time.Time `json:"-"`
	Price        int       `json:"price"`
}
