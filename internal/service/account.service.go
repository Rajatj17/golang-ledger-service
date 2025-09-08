package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	model "golang-exercise/internal/database/model"
	requestdto "golang-exercise/internal/dto/request"
	"golang-exercise/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AccountService struct {
	accRepo *repository.AccountRepository
}

func NewAccountService(accRepo *repository.AccountRepository) *AccountService {
	return &AccountService{
		accRepo: accRepo,
	}
}

const MAX_ATTEMPTS_TO_GENERATE_ACCOUNT = 5

func (accService *AccountService) generateAccountNumber(ctx context.Context, accountType model.AccountType) (string, error) {
	maxRetries := MAX_ATTEMPTS_TO_GENERATE_ACCOUNT

	for maxRetries > 0 {
		accountNumber := accService.generateUniqueAccountNumber(accountType)

		count, err := accService.accRepo.Count(ctx, accountNumber)
		if err != nil {
			return "", fmt.Errorf("failed to check uniqueness: %w", err)
		}

		if count <= 0 {
			return accountNumber, nil
		}

		fmt.Printf("Failed to generate unique account id on attempt %d", MAX_ATTEMPTS_TO_GENERATE_ACCOUNT-maxRetries+1)
		maxRetries--
	}

	return "", fmt.Errorf("failed to generate unique account id after %d attempts", MAX_ATTEMPTS_TO_GENERATE_ACCOUNT)
}

func (accService *AccountService) generateUniqueAccountNumber(accountType model.AccountType) string {
	prefix := ""
	switch accountType {
	case model.AccountTypeChecking:
		prefix = "CHE"
	case model.AccountTypeSaving:
		prefix = "SAV"
	}

	// Use current timestamp + random for uniqueness
	timestamp := time.Now().Unix()
	random := rand.Intn(9999)

	accountNumber := fmt.Sprintf("%s%d%04d", prefix, timestamp, random)

	return accountNumber
}

func (accService *AccountService) CreateAccount(ctx context.Context, req *requestdto.CreateAccount) (*model.Account, error) {
	accountNumber, err := accService.generateAccountNumber(ctx, req.AccountType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique account number: %w", err)
	}

	account := &model.Account{
		AccountNumber: accountNumber,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Balance:       req.InitialBalance,
		Currency:      req.Currency,
		AccountStatus: model.AccountActive,
		AccountType:   req.AccountType,
	}

	err = accService.accRepo.Create(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("failed to create the account")
	}

	return account, nil
}

func (accService *AccountService) GetAccount(ctx context.Context, req *requestdto.GetAccount) (*model.Account, error) {
	account, err := accService.accRepo.GetByAccountNumber(ctx, req.AccountNumber)
	if err != nil || account == nil {
		return nil, fmt.Errorf("failed to find account with such account number: %w", err)
	}

	return account, nil
}

func (accService *AccountService) UpdateBalance(ctx context.Context, accountNumber string, newBalance decimal.Decimal, tx *gorm.DB) error {
	account := &model.Account{
		Balance: newBalance,
	}

	var db *gorm.DB
	if tx != nil {
		db = tx
	} else {
		db = accService.accRepo.GetDB()
	}

	result := db.WithContext(ctx).Model(&model.Account{}).Where("account_number = ?", accountNumber).Updates(account)
	if result.Error != nil {
		return fmt.Errorf("failed to update account balance: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("failed to find account with number %s for balance update", accountNumber)
	}

	return nil
}
