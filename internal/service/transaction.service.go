package service

import (
	"context"
	"errors"
	"fmt"
	"golang-exercise/internal/database"
	model "golang-exercise/internal/database/model"

	"github.com/shopspring/decimal"
)

type TransactionService struct {
	accountService *AccountService
	txLogService   *TransactionLogService
}

type TransactionRequest struct {
	FromAccountID uint            `json:"from_account_id"`
	ToAccountID   uint            `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Description   string          `json:"description"`
	InitiatedBy   uint            `json:"initiated_by"`
}

func NewTransactionService(accountService *AccountService, txLogService *TransactionLogService) *TransactionService {
	return &TransactionService{
		accountService: accountService,
		txLogService:   txLogService,
	}
}

func (s *TransactionService) ProcessTransaction(ctx context.Context, transactionID string, accountID string, amount decimal.Decimal, transactionType model.TransactionType) error {

	// Start database transaction with pessimistic locking
	tx := database.GetPostgresDB().Begin()
	if tx.Error != nil {
		s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		}
	}()

	// Lock account and get current balance (SELECT FOR UPDATE)
	var account model.Account
	result := tx.WithContext(ctx).
		Where("account_number = ?", accountID).
		Set("gorm:query_option", "FOR UPDATE").
		First(&account)

	if result.Error != nil {
		tx.Rollback()
		s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		return fmt.Errorf("account not found or could not be locked: %w", result.Error)
	}

	var newBalance decimal.Decimal

	// Handle different transaction types with locked balance
	switch transactionType {

	case model.TransactionTypeWithdrawal:
		// Check sufficient balance for withdrawal
		if account.Balance.LessThan(amount) {
			tx.Rollback()
			s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
			return errors.New("insufficient balance")
		}

		newBalance = account.Balance.Sub(amount)

	case model.TransactionTypeDeposit:
		newBalance = account.Balance.Add(amount)

	default:
		tx.Rollback()
		s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		return errors.New("invalid transaction type")
	}

	// Update account balance within the locked transaction
	if err := s.accountService.UpdateBalance(ctx, accountID, newBalance, tx); err != nil {
		tx.Rollback()
		s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusFailed)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update transaction status to completed in MongoDB (eventual consistency)
	return s.txLogService.UpdateTransactionStatus(ctx, transactionID, model.TransactionStatusCompleted)
}
