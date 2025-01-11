package db

import (
	"api-service/config"
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool // Global connection pool

const createTableQuery = `
CREATE TABLE IF NOT EXISTS videos (
    uuid UUID PRIMARY KEY,
    title TEXT NOT NULL,
    hash TEXT NOT NULL UNIQUE,
    format TEXT NOT NULL,
    file_path TEXT NOT NULL DEFAULT '',
    uploaded_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Configure the pool
	config, err := pgxpool.ParseConfig(config.DbUrl)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v\n", err)
	}

	Pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Connected to PostgreSQL database with a connection pool.")

	_, err = Pool.Exec(ctx, createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create videos tables: %v\n", err)
	}

	log.Println("Videos tables ensured.")
}

func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("Database connection pool closed.")
	}
}
