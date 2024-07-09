package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// video 5.1
type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

// 9
type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}

// 8
type CreatedAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// video 5.5
	Password string `json:"password"`
}

type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"firstName"`
	LastName          string    `json:"lastName"`
	Number            int64     `json:"number"`
	EncryptedPassword string    `json:"-"`
	Balance           float64   `json:"balance"`
	Currency          string    `json:"currency"`
	CreatedAt         time.Time `json:"createdAt"`
}

// 7
func NewAccount(firstName, lastName, password string) (*Account, error) {
	// video 5.3
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		// ID:        rand.Intn(10000),
		FirstName:         firstName,
		LastName:          lastName,
		Number:            int64(rand.Intn(100000)),
		EncryptedPassword: string(encpw),
		CreatedAt:         time.Now().UTC(),
	}, nil
}
