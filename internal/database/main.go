package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"golang-exercise/config"
	"log"
	"time"

	"github.com/pressly/goose/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	PostgresDB *gorm.DB
	MongoDB    *mongo.Database
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func createConnectionString(dbConfig config.PostgresConfig) string {
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		dbConfig.Host,
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.Port,
	)

	return connectionString
}

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully!")
	return nil
}
func ConnectDB() {
	ConnectPostgreSQL()
	ConnectMongoDB()
}

func ConnectPostgreSQL() {
	dsn := createConnectionString(config.GetConfig().DB.Postgres)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres DB %s", err)
	}

	log.Print("Connected to the Postgres DB")

	PostgresDB = db

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Fauled to get underlying DB: %s", err)
	}
	if err := RunMigrations(sqlDB); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}
}

func ConnectMongoDB() {
	cfg := config.GetConfig()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second, // Using fixed timeout temporarily
	)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DB.Mongo.URI))
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return
	}

	MongoDB = client.Database("transaction_logs") // Use fixed database name
	log.Println("MongoDB connected successfully!")
}

func GetPostgresDB() *gorm.DB {
	return PostgresDB
}

func GetMongoDB() *mongo.Database {
	return MongoDB
}
