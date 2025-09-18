package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(cfg Config) *pgxpool.Pool {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database!")
	return pool
}
