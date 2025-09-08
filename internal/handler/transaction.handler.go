package handler

import (
	requestdto "golang-exercise/internal/dto/request"
	responsedto "golang-exercise/internal/dto/response"
	customError "golang-exercise/internal/error"
	"golang-exercise/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	accountService *service.AccountService
	txLogService   *service.TransactionLogService
}

func NewTransactionHandler(accountService *service.AccountService, txLogService *service.TransactionLogService) *TransactionHandler {
	return &TransactionHandler{
		accountService: accountService,
		txLogService:   txLogService,
	}
}

func (txHandler *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	accountNumber := c.Param("account_number")

	// Check if account exists
	req := &requestdto.GetAccount{
		AccountNumber: accountNumber,
	}

	account, err := txHandler.accountService.GetAccount(c, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", "not found in system"))
		return
	}

	// Parse query parameters
	limit := 10 // default
	offset := 0 // default

	if parsed, err := strconv.Atoi(c.Query("limit")); err == nil && parsed > 0 {
		limit = parsed
	}

	if parsed, err := strconv.Atoi(c.Query("offset")); err == nil && parsed >= 0 {
		offset = parsed
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	status := c.Query("status")

	// Get transaction history from service
	transactions, total, err := txHandler.txLogService.GetTransactionHistory(c, account.ID, limit, offset, startDate, endDate, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve transaction history",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	var historyItems []responsedto.TransactionHistoryItem
	for _, tx := range transactions {
		item := responsedto.TransactionHistoryItem{
			TransactionID: tx.TransactionId,
			FromAccountID: tx.FromAccountId,
			ToAccountID:   tx.ToAccountId,
			Amount:        tx.Amount,
			Currency:      tx.Currency,
			Type:          tx.Type,
			Status:        tx.Status,
			Memo:          tx.Memo,
			Timestamp:     tx.Timestamp,
			ProcessedAt:   tx.ProcessedAt,
		}

		historyItems = append(historyItems, item)
	}

	response := responsedto.TransactionHistoryResponse{
		Transactions: historyItems,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction history retrieved successfully",
		"data":    response,
	})
}

func (txHandler *TransactionHandler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	transaction, err := txHandler.txLogService.GetTransactionByID(c, transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("transaction", "not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction status retrieved successfully",
		"data": gin.H{
			"transaction_id": transaction.TransactionId,
			"status":         transaction.Status,
			"type":           transaction.Type,
			"amount":         transaction.Amount,
			"currency":       transaction.Currency,
			"timestamp":      transaction.Timestamp,
			"processed_at":   transaction.ProcessedAt,
		},
	})
}
