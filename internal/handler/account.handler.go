package handler

import (
	"fmt"
	"golang-exercise/internal/database/model"
	dto "golang-exercise/internal/dto"
	requestdto "golang-exercise/internal/dto/request"
	customError "golang-exercise/internal/error"
	"golang-exercise/internal/messaging"
	"time"

	"golang-exercise/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	accountService       *service.AccountService
	transactionService   *service.TransactionService
	txLogService         *service.TransactionLogService
	transactionPublisher *messaging.TransactionPublisher
}

func NewAccountHandler(accountService *service.AccountService, transactionService *service.TransactionService, txLogService *service.TransactionLogService, trxnPublisher *messaging.TransactionPublisher) *AccountHandler {
	return &AccountHandler{
		accountService:       accountService,
		transactionService:   transactionService,
		txLogService:         txLogService,
		transactionPublisher: trxnPublisher,
	}
}

var MINIMUM_ACCOUNT_BALANCE = decimal.NewFromInt(100)

func (accHandler *AccountHandler) CreateAccount(c *gin.Context) {
	var req requestdto.CreateAccount

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, customError.NewValidationError(err.Error()))
		return
	}

	account, err := accHandler.accountService.CreateAccount(c, &req)
	if err != nil {
		// Ideally we should not flag internal errors to the user, but I am doing it here so that I know if something goes wrong
		c.JSON(http.StatusInternalServerError, customError.NewInternalServerError(err.Error()))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account created successfully!",
		"data": gin.H{
			"account": account,
		},
	})
}

func (accHandler *AccountHandler) doesAccountExistsCheck(c *gin.Context, accountNumber string) *model.Account {
	req := &requestdto.GetAccount{
		AccountNumber: accountNumber,
	}

	account, err := accHandler.accountService.GetAccount(c, req)
	if err != nil {
		return nil
	}

	return account
}

func (accHandler *AccountHandler) GetAccount(c *gin.Context) {
	accountNumber := c.Param("account_number")

	account := accHandler.doesAccountExistsCheck(c, accountNumber)
	if account == nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", "not found in system"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Your account details",
		"data": gin.H{
			"account": account,
		},
	})
}

func (accHandler *AccountHandler) DepositFundsAsync(c *gin.Context) {
	var req requestdto.MoveMoneyFromAccount

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, customError.NewValidationError(err.Error()))
		return
	}

	if req.Type != model.TransactionTypeDeposit {
		c.JSON(http.StatusBadRequest, customError.NewCustomError(
			customError.ValidationError,
			"invalid transaction operation",
			"invalid operation",
		))
	}

	account := accHandler.doesAccountExistsCheck(c, req.AccountNumber)
	if account == nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", "not found in system"))
		return
	}

	// Create transaction message
	txMsg := &dto.TransactionMessage{
		ID:            uuid.New().String(),
		Type:          req.Type,
		AccountNumber: req.AccountNumber,
		Amount:        req.Amount,
		Currency:      account.Currency,
		Description:   req.Memo,
		CreatedAt:     time.Now(),
	}

	// Publish to broker instead of processing directly
	err := accHandler.transactionPublisher.PublishTransaction(txMsg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction queued successfully!",
		"data": gin.H{
			"TransactionID": txMsg.ID,
			"Status":        "IN_PROGRESS",
		},
	})
}

func (accHandler *AccountHandler) WithdrawFundsAsync(c *gin.Context) {
	var req requestdto.MoveMoneyFromAccount

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, customError.NewValidationError(err.Error()))
		return
	}

	if req.Type != model.TransactionTypeWithdrawal {
		c.JSON(http.StatusBadRequest, customError.NewCustomError(
			customError.ValidationError,
			"invalid transaction operation",
			"invalid operation",
		))
	}

	account := accHandler.doesAccountExistsCheck(c, req.AccountNumber)
	if account == nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", "not found in system"))
		return
	}

	if account.Balance.Sub(req.Amount).LessThan(MINIMUM_ACCOUNT_BALANCE) {
		c.JSON(http.StatusBadRequest, customError.NewValidationError("insufficient balance"))
	}

	// Create transaction message
	txMsg := &dto.TransactionMessage{
		ID:            uuid.New().String(),
		Type:          req.Type,
		AccountNumber: req.AccountNumber,
		Amount:        req.Amount,
		Currency:      account.Currency,
		Description:   req.Memo,
		CreatedAt:     time.Now(),
	}

	// Publish to broker instead of processing directly
	err := accHandler.transactionPublisher.PublishTransaction(txMsg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Withdraw successfully",
		"data": gin.H{
			"TransactionID": fmt.Sprintf("TXN_%d", time.Now().UnixNano()),
		},
	})
}

func (accHandler *AccountHandler) ProcessFunds(c *gin.Context) {
	var req requestdto.MoveMoneyFromAccount

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, customError.NewValidationError(err.Error()))
		return
	}

	// Validate transaction type
	if req.Type != model.TransactionTypeDeposit && req.Type != model.TransactionTypeWithdrawal {
		c.JSON(http.StatusBadRequest, customError.NewCustomError(
			customError.ValidationError,
			"invalid transaction operation",
			"invalid operation",
		))
		return
	}

	// Check if account exists
	account := accHandler.doesAccountExistsCheck(c, req.AccountNumber)
	if account == nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", "not found in system"))
		return
	}

	// For withdrawal, check minimum balance
	if req.Type == model.TransactionTypeWithdrawal {
		if account.Balance.Sub(req.Amount).LessThan(MINIMUM_ACCOUNT_BALANCE) {
			c.JSON(http.StatusBadRequest, customError.NewValidationError("insufficient balance"))
			return
		}
	}

	// Generate transaction ID
	transactionID := fmt.Sprintf("TXN_%d", time.Now().UnixNano())

	// Create transaction log entry with IN_PROGRESS status
	txLog := &model.TransactionLog{
		TransactionId: transactionID,
		FromAccountId: account.ID,
		ToAccountId:   account.ID, // Same account for single account operations
		Amount:        req.Amount,
		Currency:      account.Currency,
		Type:          req.Type,
		Status:        model.TransactionStatusInprogress,
		Memo:          req.Memo,
		InitiatedBy:   1, // TODO: Using the userid from jwt when auth is enabled
		Timestamp:     time.Now(),
	}

	if err := accHandler.txLogService.LogTransaction(c, txLog); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create transaction log",
			"details": err.Error(),
		})
		return
	}

	// Create transaction message for RabbitMQ
	txMsg := &dto.TransactionMessage{
		ID:            transactionID,
		Type:          req.Type,
		AccountNumber: req.AccountNumber,
		Amount:        req.Amount,
		Currency:      account.Currency,
		Description:   req.Memo,
		CreatedAt:     time.Now(),
	}

	// Publish to RabbitMQ broker instead of processing directly
	err := accHandler.transactionPublisher.PublishTransaction(txMsg)
	if err != nil {
		// Update transaction log to failed status
		accHandler.txLogService.UpdateTransactionStatus(c, transactionID, model.TransactionStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to queue transaction",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Transaction queued successfully!",
		"data": gin.H{
			"TransactionID": transactionID,
			"Status":        "IN_PROGRESS",
		},
	})
}

func (accHandler *AccountHandler) GetAccountBalance(c *gin.Context) {
	accountNumber := c.Param("account_number")
	req := &requestdto.GetAccount{
		AccountNumber: accountNumber,
	}

	account, err := accHandler.accountService.GetAccount(c, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, customError.NewEntityNotFoundError("account", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Your account balance",
		"data": gin.H{
			"Balance": account.Balance,
		},
	})
}
