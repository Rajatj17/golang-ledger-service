package model

import (
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
)

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusInprogress TransactionStatus = "IN_PROGRESS"
	TransactionStatusCompleted  TransactionStatus = "COMPLETED"
	TransactionStatusFailed     TransactionStatus = "FAILED"
)

type TransactionLog struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionId string             `bson:"transaction_id" json:"transaction_id"`
	FromAccountId uint               `bson:"from_account_id" json:"from_account_id"`
	ToAccountId   uint               `bson:"to_account_id" json:"to_account_id"`
	Amount        decimal.Decimal    `bson:"amount" json:"amount"`
	Currency      string             `bson:"currency" json:"currency"`
	Type          TransactionType    `bson:"type" json:"type"`
	Status        TransactionStatus  `bson:"status" json:"status"`
	Memo          string             `bson:"memo" json:"description"`
	Metadata      map[string]any     `bson:"metadata" json:"metadata"`
	Timestamp     time.Time          `bson:"timestamp" json:"timestamp"`
	ProcessedAt   *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`

	// For audit trail
	InitiatedBy uint `bson:"initiated_by" json:"initiated_by"`
	RetryCount  int  `bson:"retry_count" json:"retry_count"`
}
