package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
)

type PostgresStorage struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresStorage() (*PostgresStorage, error) {
	dsn := os.Getenv("PG_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Errorf("Error opening DB")
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		log.Fatal("Ошибка при проверке соединения с базой данных: %v", err)
	}
	postgresStorage := PostgresStorage{
		db:  db,
		ctx: context.Background(),
	}
	return &postgresStorage, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
