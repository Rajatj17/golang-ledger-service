package repository

import (
	"context"
	"fmt"
	"golang-exercise/internal/database"
	"golang-exercise/internal/database/model"

	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{
		db: database.GetPostgresDB(),
	}
}

func NewAccountRepositoryWithDB(db *gorm.DB) *AccountRepository {
	return &AccountRepository{
		db: db,
	}
}

func (repo *AccountRepository) GetDB() *gorm.DB {
	return repo.db
}

func (repo *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	result := repo.db.WithContext(ctx).Create(&account)
	if result.Error != nil {
		return fmt.Errorf("failed to create the user: %w", result.Error)
	}

	return result.Error
}

func (repo *AccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*model.Account, error) {
	account := &model.Account{}
	result := repo.db.WithContext(ctx).First(account, "account_number = ?", accountNumber)

	if result.Error != nil {
		return nil, result.Error
	}

	return account, nil
}

func (repo *AccountRepository) Count(ctx context.Context, accountNumber string) (int64, error) {
	var count int64

	result := repo.db.WithContext(ctx).Model(&model.Account{}).Where("account_number = ?", accountNumber).Count(&count)

	return count, result.Error
}

func (repo *AccountRepository) Update(ctx context.Context, accountNumber string, account *model.Account) error {
	result := repo.db.WithContext(ctx).Model(&model.Account{}).Where("account_number = ?", accountNumber).Updates(account)
	if result.Error != nil {
		return fmt.Errorf("failed to update the user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("failed to find user with id %s for update: %w", accountNumber, result.Error)
	}

	return nil
}

func (repo *AccountRepository) GetAll() ([]*model.Account, error) {
	// Logic for getting all accounts, can be used by the admins API
	return nil, nil
}

func (repo *AccountRepository) Delete() error {
	// Logic for deleting the account, if needed
	return nil
}
