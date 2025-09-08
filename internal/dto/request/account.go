package requestdto

import (
	"golang-exercise/internal/database/model"

	"github.com/shopspring/decimal"
)

type CreateAccount struct {
	FirstName      string            `json:"first_name" validate:"required"`
	LastName       string            `json:"last_name" validate:"required"`
	AccountType    model.AccountType `json:"account_type" validate:"required"`
	Currency       string            `json:"currency" validate:"required"`
	InitialBalance decimal.Decimal   `json:"initial_balance" validate:"required"`
}

type GetAccount struct {
	AccountNumber string `json:"account_number" validate:"required"`
}

type UpdateAccountStatus struct {
	Status model.AccountStatus `json:"status"`
}

type MoveMoneyFromAccount struct {
	AccountNumber string                `json:"account_number" validate:"required"`
	Amount        decimal.Decimal       `json:"amount" validate:"required"`
	Type          model.TransactionType `json:"type" validate:"required"`
	Memo          string                `json:"memo"`
}

type GetTransactionHistory struct {
	AccountNumber string `json:"account_number" validate:"required"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	StartDate     string `json:"start_date,omitempty"` // Format: YYYY-MM-DD
	EndDate       string `json:"end_date,omitempty"`   // Format: YYYY-MM-DD
	Status        string `json:"status,omitempty"`
}
