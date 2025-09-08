package integration

import (
	"context"
	"testing"

	"golang-exercise/internal/database/model"
	"golang-exercise/internal/repository"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite
	db          *gorm.DB
	accountRepo *repository.AccountRepository
}

func (suite *RepositoryTestSuite) SetupSuite() {
	// Setup test database connection
	// In a real scenario, you'd use a test database or Docker container
	dsn := "host=localhost user=postgres password=postgres dbname=banking_ledger_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		suite.T().Skip("Skipping integration tests: PostgreSQL not available")
		return
	}

	// Auto-migrate models
	err = db.AutoMigrate(&model.Account{})
	if err != nil {
		suite.T().Fatal("Failed to migrate test database:", err)
	}

	suite.db = db
	suite.accountRepo = repository.NewAccountRepositoryWithDB(db)
}

func (suite *RepositoryTestSuite) SetupTest() {
	// Clean up test data before each test
	suite.db.Where("1 = 1").Delete(&model.Account{})
}

func (suite *RepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		// Clean up database
		suite.db.Where("1 = 1").Delete(&model.Account{})
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *RepositoryTestSuite) TestAccountRepository_Create() {
	ctx := context.Background()

	account := &model.Account{
		AccountNumber: "TEST123456",
		FirstName:     "Rajat",
		LastName:      "J",
		Balance:       decimal.NewFromInt(1000),
		Currency:      "INR",
		AccountType:   model.AccountTypeChecking,
		AccountStatus: model.AccountActive,
	}

	err := suite.accountRepo.Create(ctx, account)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), account.ID)

	// Verify account was created
	var retrievedAccount model.Account
	err = suite.db.Where("account_number = ?", "TEST123456").First(&retrievedAccount).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), account.FirstName, retrievedAccount.FirstName)
	assert.Equal(suite.T(), account.LastName, retrievedAccount.LastName)
	assert.True(suite.T(), account.Balance.Equal(retrievedAccount.Balance))
}

func (suite *RepositoryTestSuite) TestAccountRepository_GetByAccountNumber() {
	ctx := context.Background()

	// Create test account
	account := &model.Account{
		AccountNumber: "TEST123457",
		FirstName:     "Foo",
		LastName:      "Bar",
		Balance:       decimal.NewFromInt(2000),
		Currency:      "USD",
		AccountType:   model.AccountTypeSaving,
		AccountStatus: model.AccountActive,
	}

	err := suite.accountRepo.Create(ctx, account)
	assert.NoError(suite.T(), err)

	// Retrieve account
	retrievedAccount, err := suite.accountRepo.GetByAccountNumber(ctx, "TEST123457")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedAccount)
	assert.Equal(suite.T(), "Foo", retrievedAccount.FirstName)
	assert.Equal(suite.T(), "Bar", retrievedAccount.LastName)
	assert.True(suite.T(), decimal.NewFromInt(2000).Equal(retrievedAccount.Balance))
}

func (suite *RepositoryTestSuite) TestAccountRepository_GetByAccountNumber_NotFound() {
	ctx := context.Background()

	retrievedAccount, err := suite.accountRepo.GetByAccountNumber(ctx, "NONEXISTENT123")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), retrievedAccount)
}

func (suite *RepositoryTestSuite) TestAccountRepository_Count() {
	ctx := context.Background()

	// Count should be 0 initially
	count, err := suite.accountRepo.Count(ctx, "UNIQUE123")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)

	// Create account
	account := &model.Account{
		AccountNumber: "UNIQUE123",
		FirstName:     "Test",
		LastName:      "User",
		Balance:       decimal.NewFromInt(500),
		Currency:      "USD",
		AccountType:   model.AccountTypeChecking,
		AccountStatus: model.AccountActive,
	}

	err = suite.accountRepo.Create(ctx, account)
	assert.NoError(suite.T(), err)

	// Count should be 1 now
	count, err = suite.accountRepo.Count(ctx, "UNIQUE123")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *RepositoryTestSuite) TestAccountRepository_ConcurrentTransactions() {
	ctx := context.Background()

	// Create test account
	account := &model.Account{
		AccountNumber: "CONCURRENT123",
		FirstName:     "Concurrent",
		LastName:      "Test",
		Balance:       decimal.NewFromInt(1000),
		Currency:      "USD",
		AccountType:   model.AccountTypeChecking,
		AccountStatus: model.AccountActive,
	}

	err := suite.accountRepo.Create(ctx, account)
	assert.NoError(suite.T(), err)

	// Test pessimistic locking with SELECT FOR UPDATE
	tx1 := suite.db.Begin()
	tx2 := suite.db.Begin()

	// First transaction locks the account
	var lockedAccount1 model.Account
	err1 := tx1.WithContext(ctx).
		Where("account_number = ?", "CONCURRENT123").
		Set("gorm:query_option", "FOR UPDATE").
		First(&lockedAccount1).Error
	assert.NoError(suite.T(), err1)

	// Update balance in first transaction
	newBalance := lockedAccount1.Balance.Add(decimal.NewFromInt(100))
	err1 = tx1.Model(&lockedAccount1).Update("balance", newBalance).Error
	assert.NoError(suite.T(), err1)

	// Commit first transaction
	err1 = tx1.Commit().Error
	assert.NoError(suite.T(), err1)

	// Second transaction should see updated balance
	var account2 model.Account
	err2 := tx2.WithContext(ctx).
		Where("account_number = ?", "CONCURRENT123").
		First(&account2).Error
	assert.NoError(suite.T(), err2)

	tx2.Commit()

	// Final balance should be 1100
	finalAccount, err := suite.accountRepo.GetByAccountNumber(ctx, "CONCURRENT123")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), decimal.NewFromInt(1100).Equal(finalAccount.Balance))
}

func TestRepositoryIntegrationSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
