package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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
	CustomerID    int       `json:"customer_id"`
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
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	TransactionStatusID int       `json:"transaction_status_id"`
	TimeCreated         time.Time `json:"-"`
	TimeUpdated         time.Time `json:"-"`
	CustomerID          int       `json:"customer_id"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
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

type Customer struct {
	ID          int       `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	TimeCreated time.Time `json:"-"`
	TimeUpdated time.Time `json:"-"`
}

func (m *DBModel) GetTea(id int) (Teas, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var tea Teas

	row := m.DB.QueryRowContext(ctx, `select 
	id, name, description, inventory_level, price, coalesce(image,''),
	created_at, updated_at
	from 
	teas
	where id = ?`, id)
	err := row.Scan(
		&tea.ID,
		&tea.Name,
		&tea.Description,
		&tea.InventoryAmt,
		&tea.Price,
		&tea.Image,
		&tea.TimeCreated,
		&tea.TimeUpdated,
	)
	if err != nil {
		return tea, err
	}

	return tea, nil
}

// returns all teas from the DB
func (m *DBModel) AllTeas() ([]Teas, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, name, description, inventory_level, price,
		       COALESCE(image,''), created_at, updated_at
		FROM   teas
		ORDER BY name
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teas []Teas
	for rows.Next() {
		var t Teas
		err = rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.InventoryAmt,
			&t.Price, &t.Image, &t.TimeCreated, &t.TimeUpdated,
		)
		if err != nil {
			return nil, err
		}
		teas = append(teas, t)
	}

	return teas, nil
}

func (m *DBModel) DecrementInventory(tx *sql.Tx, teaID, qty int) error {
	res, err := tx.Exec(`
		UPDATE teas
		SET    inventory_level = inventory_level - ?
		WHERE  id = ? AND inventory_level >= ?`,
		qty, teaID, qty,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("not enough stock")
	}
	return nil
}

// inserts a new txn in the DB and returns its ID
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into transactions
			(amount, currency, last_four, bank_return_code, expiry_month, expiry_year,
			payment_intent, payment_method, transaction_status_id, created_at, updated_at)
			values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,

		txn.ExpiryYear,

		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return 0, err

	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// inserts a new order and returns it's ID
func (m *DBModel) InsertOrder(order Order) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into orders
			(tea_id, transaction_id, status_id, quantity, customer_id,
			amount, created_at, updated_at)
			values (?, ?, ?, ?, ?, ?, ?, ?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		order.TeaID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.CustomerID,
		order.Amount,

		time.Now(),
		time.Now(),
	)

	if err != nil {
		return 0, err

	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *DBModel) InsertCustomer(c Customer) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		insert into customers
			(first_name, last_name, email, created_at, updated_at)
			values (?, ?, ?, ?, ?)
			`
	result, err := m.DB.ExecContext(ctx, stmt,
		c.FirstName,
		c.LastName,
		c.Email,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return 0, err

	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *DBModel) InsertOrderTx(tx *sql.Tx, order Order) (int, error) {
	stmt := `
		INSERT INTO orders
		    (tea_id, transaction_id, status_id, quantity, customer_id,
		     amount, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := tx.Exec(stmt,
		order.TeaID, order.TransactionID, order.StatusID,
		order.Quantity, order.CustomerID, order.Amount,
		time.Now(), time.Now(),
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

// Checks DB to see if a user with input email exists,
// returns the user if it exists, returns error otherwise
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)

	var u User

	row := m.DB.QueryRowContext(ctx, `
		select
			id, first_name, last_name, email, password, created_at, updated_at
		from
			users
		where email = ?
		`, email)

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.TimeCreated,
		&u.TimeUpdated,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) LatestTeas(n int) ([]*Teas, error) {
	rows, err := m.DB.QueryContext(context.Background(),
		`SELECT id, name, description, price, image, created_at
         FROM teas
         ORDER BY created_at DESC
         LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teas []*Teas
	for rows.Next() {
		var t Teas
		if err := rows.Scan(&t.ID, &t.Name, &t.Description,
			&t.Price, &t.Image, &t.TimeCreated); err != nil {
			return nil, err
		}
		teas = append(teas, &t)
	}
	return teas, rows.Err()
}
