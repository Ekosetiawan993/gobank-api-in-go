package main

import (
	"database/sql"

	"fmt"

	_ "github.com/lib/pq"
)

// 1
type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

// 2
func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=gotest password=tatapjang sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

// 5
func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

// 4
func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS account (
		id serial primary key,
		first_name VARCHAR(50),
		last_name VARCHAR(50),
		number serial,
		balance NUMERIC(15,3),
		currency TEXT CHECK (currency IN ('USD', 'IDR')),
		created_at timestamp
		)
	`
	_, err := s.db.Exec(query)
	return err
}

// 3. implementing interface so me main Run can executed
func (s *PostgresStore) CreateAccount(acc *Account) error {
	insertStatement := `
	INSERT INTO account (first_name, last_name, number, balance, currency, created_at)
	VALUES
	($1, $2, $3, $4, $5, $6)
	`
	resp, err := s.db.Exec(
		insertStatement,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.Balance,
		"USD",
		acc.CreatedAt)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", resp)
	return nil
}
func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Exec("DELETE FROM account WHERE id = $1", id)
	return err
}
func (s *PostgresStore) UpdateAccount(*Account) error {
	return nil
}
func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	query := `SELECT * FROM account WHERE id = $1`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("Account with id %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	getStatements := `
	SELECT * FROM account ORDER BY id
	`
	rows, err := s.db.Query(getStatements)
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.Currency,
		&account.CreatedAt)

	return account, err
}
