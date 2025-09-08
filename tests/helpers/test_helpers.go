package helpers

import (
	"context"
	"testing"

	"golang-exercise/internal/database/model"
	"golang-exercise/internal/repository"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDatabase provides utilities for database testing
type TestDatabase struct {
	DB *gorm.DB
}

// NewTestDatabase creates a new test database connection
func NewTestDatabase() (*TestDatabase, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=banking_ledger_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate test models
	err = db.AutoMigrate(&model.Account{})
	if err != nil {
		return nil, err
	}

	return &TestDatabase{DB: db}, nil
}

// CleanupTestData removes all test data from the database
func (td *TestDatabase) CleanupTestData() {
	td.DB.Where("1 = 1").Delete(&model.Account{})
}

// CreateTestAccount creates a test account with default values
func (td *TestDatabase) CreateTestAccount(accountNumber string) *model.Account {
	account := &model.Account{
		AccountNumber: accountNumber,
		FirstName:     "Test",
		LastName:      "User",
		Balance:       decimal.NewFromInt(1000),
		Currency:      "USD",
		AccountType:   model.AccountTypeChecking,
		AccountStatus: model.AccountActive,
	}

	repo := repository.NewAccountRepository()
	err := repo.Create(context.Background(), account)
	if err != nil {
		return nil
	}

	return account
}

// Close closes the database connection
func (td *TestDatabase) Close() {
	if td.DB != nil {
		sqlDB, _ := td.DB.DB()
		sqlDB.Close()
	}
}

// AssertAccountBalance verifies an account has the expected balance
func AssertAccountBalance(t *testing.T, db *gorm.DB, accountNumber string, expectedBalance decimal.Decimal) {
	repo := repository.NewAccountRepository()
	account, err := repo.GetByAccountNumber(context.Background(), accountNumber)

	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.True(t, expectedBalance.Equal(account.Balance),
		"Expected balance %s, got %s", expectedBalance.String(), account.Balance.String())
}

// AssertAccountExists verifies an account exists
func AssertAccountExists(t *testing.T, db *gorm.DB, accountNumber string) {
	repo := repository.NewAccountRepository()
	account, err := repo.GetByAccountNumber(context.Background(), accountNumber)

	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, accountNumber, account.AccountNumber)
}

// AssertAccountNotExists verifies an account does not exist
func AssertAccountNotExists(t *testing.T, db *gorm.DB, accountNumber string) {
	repo := repository.NewAccountRepository()
	account, err := repo.GetByAccountNumber(context.Background(), accountNumber)

	assert.Error(t, err)
	assert.Nil(t, account)
}
