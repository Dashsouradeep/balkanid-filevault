package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func ConnectDB(cfg Config) *pgx.Conn {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v\n", err)
	}

	fmt.Println("✅ Connected to PostgreSQL database!")
	return conn
}
