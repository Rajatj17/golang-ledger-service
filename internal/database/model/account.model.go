package model

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AccountStatus string

const (
	AccountActive AccountStatus = "ACTIVE"
	AccountFrozen AccountStatus = "FROZEN"
	AccountClosed AccountStatus = "CLOSED"
)

type AccountType string

const (
	AccountTypeChecking AccountType = "CHECKING"
	AccountTypeSaving   AccountType = "SAVINGS"
)

type Account struct {
	gorm.Model
	AccountNumber string
	FirstName     string
	LastName      string
	Balance       decimal.Decimal
	Currency      string
	AccountType   AccountType
	AccountStatus AccountStatus
}
