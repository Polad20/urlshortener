package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDatabaseDSN string

func TestMain(m *testing.M) {
	testDatabaseDSN = os.Getenv("TEST_BASE_DSN")
	if testDatabaseDSN == "" {
		fmt.Println("Перменная окружения TEST_BASE_DSN не задана - тесты не будут запущены")
		os.Exit(0)
	}
	code := m.Run()
	os.Exit(code)
}

func openTestDBConnection(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", testDatabaseDSN)
	require.NoError(t, err, "не удалось подключиться к тестовой БД")
	err = db.Ping()
	require.NoError(t, err, "Не удалось проверить соединение с тестовой БД")
	return db
}

func cleanTables(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM public.test_table; DELETE FROM urls;")
	require.NoError(t, err, "Не удалось очистить тестовые таблицы")
}

func TestFindUsersOrigURL(t *testing.T) {
	db := openTestDBConnection(t)
	defer db.Close()

	tests := []struct {
		name                string
		setupData           func(t *testing.T, tx *sql.Tx)
		userID              string
		shortURL            string
		expectedOriginalURL string
		expectedError       error
	}{
		{
			name: "Succesfull find",
			setupData: func(t *testing.T, tx *sql.Tx) {
				_, err := tx.Exec(`INSERT INTO public.test_table (UserID, Correlation_id, Original_url, Short_url) VALUES ($1, $2, $3, $4)`,
					"user1", "corr1", "https://example.com/original1", "http://localhost:8080/short1")
				require.NoError(t, err, "Не удалось вставить тестовые данные для кейса 'Succesfull find'")
			},
			userID:              "user1",
			shortURL:            "http://localhost:8080/short1",
			expectedOriginalURL: "https://example.com/original1",
			expectedError:       nil,
		},
		{
			name: "URL not found for user",
			setupData: func(t *testing.T, tx *sql.Tx) {
				_, err := tx.Exec(`INSERT INTO public.test_table (UserID, Correlation_id, Original_url, Short_url) VALUES ($1, $2, $3, $4)`,
					"user2", "corr2a", "https://example.com/original2a", "http://localhost:8080/short2a")
				require.NoError(t, err, "Не удалось вставить тестовые данные для кейса 'URL not found for user'")
			},
			userID:              "user2",
			shortURL:            "http://localhost:8080/short2b",
			expectedOriginalURL: "",
			expectedError:       ErrURLNotFoundForUser,
		},
		{
			name:                "User not found test",
			setupData:           func(t *testing.T, tx *sql.Tx) {},
			userID:              "nope",
			shortURL:            "http://localhost:8080/nope",
			expectedOriginalURL: "",
			expectedError:       ErrURLNotFoundForUser,
		},
		{
			name:                "No userID provided test",
			setupData:           func(t *testing.T, tx *sql.Tx) {},
			userID:              "",
			shortURL:            "http://localhost:8080/any",
			expectedOriginalURL: "",
			expectedError:       errors.New("Got Empty userID"),
		},
		{
			name:                "No shortURL provided",
			setupData:           func(t *testing.T, tx *sql.Tx) {},
			userID:              "user11",
			shortURL:            "",
			expectedOriginalURL: "",
			expectedError:       errors.New("Got empty shortURL"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := db.BeginTx(context.Background(), nil)
			require.NoError(t, err, "Не удалось начать транзакцию для подтеста")
			defer tx.Rollback()
			cleanTables(t, db)
			tc.setupData(t, tx)
			pgOne := PostgresStorage{DB: db}
			foundURL, err := pgOne.FindUsersOrigURL(tc.userID, tc.shortURL)
			if tc.expectedError == nil {
				assert.NoError(t, err, fmt.Sprintf("Ожидали nil, но получили %v", err))
			} else {
				assert.Error(t, err, "Ожидали ошибку, получили nil")
				if tc.expectedError == ErrURLNotFoundForUser {
					assert.ErrorIs(t, err, ErrURLNotFoundForUser, fmt.Sprintf("Ожидали ErrURLNotFoundForUser, а получили %v", err))
				} else {
					assert.Equal(t, tc.expectedError.Error(), err.Error(), fmt.Sprintf("Ожидали текст ошибки '%s', получили '%s'", tc.expectedError.Error(), err.Error()))
				}
			}
			if err == nil {
				assert.Equal(t, tc.expectedOriginalURL, foundURL, "Найденный URL не совпадает")
			} else {
				assert.Equal(t, "", foundURL, "При ошибке найденный оригинальный URL должен быть пустой строкой")
			}
		})
	}
}
