package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Polad20/urlshortener/internal/model"
	"github.com/Polad20/urlshortener/internal/shortener"
)

var ErrURLNotFoundForUser = errors.New("URL not found for user")

type PostgresStorage struct {
	DB *sql.DB
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
		DB: db,
	}
	return &postgresStorage, nil
}

func (p *PostgresStorage) BaseSave(ctx context.Context, dbToSave []model.DbSave) error {
	tx, err := p.DB.BeginTx(ctx, nil)
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
	return p.DB.PingContext(ctx)
}

func (p *PostgresStorage) GetURLsByUser(userID string) ([]model.ShortenedURL, error) {
	return nil, nil
}

func (p *PostgresStorage) SaveURL(userID, shortURL, originalURL string) error {
	return nil
}

func (p *PostgresStorage) FindUsersOrigURL(userID, shortURL string) (string, error) {
	if userID == "" || shortURL == "" {
		return "", errors.New("Got empty userID or shortURL")
	}
	query := "SELECT Original_url FROM public.test_table WHERE UserID = $1 AND Short_url = $2"
	var originalURL string
	row := p.DB.QueryRow(query, userID, shortURL)
	err := row.Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("short URL %s not found for user %s: %w", shortURL, userID, ErrURLNotFoundForUser)
		}
		return "", fmt.Errorf("failed to scan result for short URL %s for user %s: %w", shortURL, userID, err)
	}
	return originalURL, nil
}

func (p *PostgresStorage) DeleteURLs(userID string, batchIDs []string, tx *sql.Tx) error {
	if len(batchIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(batchIDs))
	args := make([]any, len(batchIDs)+1)
	args[0] = userID
	for i := 0; i < len(batchIDs); i++ {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = batchIDs[i]
	}
	query := `UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND id IN (` +
		strings.Join(placeholders, ", ") + `);`

	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("Failed to execute batch update")
	}
	return nil
}
