package responsedto

import (
	"golang-exercise/internal/database/model"
	"time"

	"github.com/shopspring/decimal"
)

type CreateAccountResponse struct {
	Account *model.Account
}

type TransferResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	Message       string `json:"message"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionHistoryItem `json:"transactions"`
	Total        int                      `json:"total"`
	Limit        int                      `json:"limit"`
	Offset       int                      `json:"offset"`
}

type TransactionHistoryItem struct {
	TransactionID     string                  `json:"transaction_id"`
	FromAccountID     uint                    `json:"from_account_id,omitempty"`
	ToAccountID       uint                    `json:"to_account_id,omitempty"`
	FromAccountNumber string                  `json:"from_account_number,omitempty"`
	ToAccountNumber   string                  `json:"to_account_number,omitempty"`
	Amount            decimal.Decimal         `json:"amount"`
	Currency          string                  `json:"currency"`
	Type              model.TransactionType   `json:"type"`
	Status            model.TransactionStatus `json:"status"`
	Memo              string                  `json:"memo"`
	Timestamp         time.Time               `json:"timestamp"`
	ProcessedAt       *time.Time              `json:"processed_at,omitempty"`
}

type AccountBalanceResponse struct {
	AccountNumber string          `json:"account_number"`
	Balance       decimal.Decimal `json:"balance"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
}
