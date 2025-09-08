package dto

import (
	"golang-exercise/internal/database/model"
	"time"

	"github.com/shopspring/decimal"
)

type TransactionMessage struct {
	ID              string                `json:"id"`
	Type            model.TransactionType `json:"type"`
	AccountNumber   string                `json:"account_number"`
	ToAccountNumber string                `json:"to_account_number,omitempty"` // for transfers to another account
	Amount          decimal.Decimal       `json:"amount"`
	Currency        string                `json:"currency"`
	Description     string                `json:"description,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
}
