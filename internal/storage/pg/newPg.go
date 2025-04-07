package pg

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/Polad20/urlshortener/internal/model"
	"github.com/Polad20/urlshortener/internal/shortener"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	dsn := os.Getenv("PG_URL")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Error opening DB, %v", err)
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		log.Fatalf("Ошибка при проверке соединения с базой данных: %v", err)
	}
	postgresStorage := PostgresStorage{
		db: db,
	}
	return &postgresStorage, nil
}

func (p *PostgresStorage) BaseSave(ctx context.Context, dbToSave []model.DbSave) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error creating transaction: %v", err)
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO public.test_table(UserID, Correlation_id, Original_url,Short_url) VALUES($1,$2,$3,$4) ON CONFLICT(Original_url) DO NOTHING")
	if err != nil {
		log.Printf("Error creating statement: %v", err)
		return err
	}
	defer stmt.Close()
	for _, v := range dbToSave {
		if _, err = stmt.ExecContext(ctx, v.UserID, v.Correlation_id, v.Original_url, v.Short_url); err != nil {
			log.Printf("Error execing statement: %v", err)
			return err
		}
	}
	return tx.Commit()

}

func DbSavePrepare(userID string, item model.Incoming, shortener *shortener.Shortener) (model.DbSave, error) {
	var newDbItem model.DbSave
	newDbItem.UserID = userID
	newDbItem.Correlation_id = item.Correlation_id
	newDbItem.Original_url = item.Original_url
	newDbItem.Short_url = shortener.Shorten()
	return newDbItem, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *PostgresStorage) GetURLsByUser(userID string) ([]model.ShortenedURL, error) {
	return nil, nil
}

func (p *PostgresStorage) SaveURL(userID, shortURL, originalURL string) error {
	return nil
}
