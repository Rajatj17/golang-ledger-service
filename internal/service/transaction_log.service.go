package service

import (
	"context"

	"golang-exercise/internal/database/model"
	"golang-exercise/internal/repository"
)

type TransactionLogService struct {
	txLogRepo *repository.TransactionLogRepository
}

func NewTransactionLogService(txLogRepo *repository.TransactionLogRepository) *TransactionLogService {
	return &TransactionLogService{
		txLogRepo: txLogRepo,
	}
}

func (s *TransactionLogService) LogTransaction(ctx context.Context, txLog *model.TransactionLog) error {
	return s.txLogRepo.Create(ctx, txLog)
}

func (s *TransactionLogService) GetTransactionsByAccount(ctx context.Context, accountID uint, limit int64) ([]model.TransactionLog, error) {
	return s.txLogRepo.GetByAccountID(ctx, accountID, limit)
}

func (s *TransactionLogService) GetTransactionByID(ctx context.Context, transactionID string) (*model.TransactionLog, error) {
	return s.txLogRepo.GetByTransactionID(ctx, transactionID)
}

func (s *TransactionLogService) UpdateTransactionStatus(ctx context.Context, transactionID string, status model.TransactionStatus) error {
	return s.txLogRepo.UpdateStatus(ctx, transactionID, status)
}

func (s *TransactionLogService) GetTransactionsByStatus(ctx context.Context, status string, limit int64) ([]model.TransactionLog, error) {
	return s.txLogRepo.GetByStatus(ctx, status, limit)
}

func (s *TransactionLogService) GetAllTransactions(ctx context.Context, limit int64, offset int64) ([]model.TransactionLog, error) {
	return s.txLogRepo.GetAll(ctx, limit, offset)
}

func (s *TransactionLogService) GetTransactionHistory(ctx context.Context, accountID uint, limit int, offset int, startDate, endDate, status string) ([]model.TransactionLog, int, error) {
	return s.txLogRepo.GetTransactionHistory(ctx, accountID, limit, offset, startDate, endDate, status)
}
