package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"golang-exercise/internal/database"
	"golang-exercise/internal/database/model"
)

type TransactionLogRepository struct {
	collection *mongo.Collection
}

func NewTransactionLogRepository() *TransactionLogRepository {
	mongoDB := database.GetMongoDB()
	return &TransactionLogRepository{
		collection: mongoDB.Collection("transaction_logs"),
	}
}

func (repo *TransactionLogRepository) Create(ctx context.Context, txLog *model.TransactionLog) error {
	txLog.ID = primitive.NewObjectID()
	txLog.Timestamp = time.Now()

	_, err := repo.collection.InsertOne(ctx, txLog)
	return err
}

func (repo *TransactionLogRepository) GetByAccountID(ctx context.Context, accountID uint, limit int64) ([]model.TransactionLog, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_account_id": accountID},
			{"to_account_id": accountID},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cursor, err := repo.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.TransactionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (repo *TransactionLogRepository) GetByTransactionID(ctx context.Context, transactionID string) (*model.TransactionLog, error) {
	filter := bson.M{"transaction_id": transactionID}

	var txLog model.TransactionLog
	err := repo.collection.FindOne(ctx, filter).Decode(&txLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &txLog, nil
}

func (repo *TransactionLogRepository) UpdateStatus(ctx context.Context, transactionID string, status model.TransactionStatus) error {
	filter := bson.M{"transaction_id": transactionID}
	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"processed_at": time.Now(),
		},
	}

	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *TransactionLogRepository) GetByStatus(ctx context.Context, status string, limit int64) ([]model.TransactionLog, error) {
	filter := bson.M{"status": status}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(limit)
	cursor, err := repo.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.TransactionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (repo *TransactionLogRepository) GetAll(ctx context.Context, limit int64, offset int64) ([]model.TransactionLog, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := repo.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []model.TransactionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func (repo *TransactionLogRepository) GetTransactionHistory(ctx context.Context, accountID uint, limit int, offset int, startDate, endDate, status string) ([]model.TransactionLog, int, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_account_id": accountID},
			{"to_account_id": accountID},
		},
	}

	// Add date filter if provided
	if startDate != "" || endDate != "" {
		dateFilter := bson.M{}
		if startDate != "" {
			if startTime, err := time.Parse("2006-01-02", startDate); err == nil {
				dateFilter["$gte"] = startTime
			}
		}
		if endDate != "" {
			if endTime, err := time.Parse("2006-01-02", endDate); err == nil {
				dateFilter["$lte"] = endTime.Add(24 * time.Hour) // End of day
			}
		}
		if len(dateFilter) > 0 {
			filter["timestamp"] = dateFilter
		}
	}

	// Add status filter if provided
	if status != "" {
		filter["status"] = status
	}

	// Get total count
	totalCount, err := repo.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := repo.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []model.TransactionLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, int(totalCount), nil
}
