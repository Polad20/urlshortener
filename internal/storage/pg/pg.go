package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	errNotUnique         = errors.New("not unique original url")
	errNotUniqueShortUrl = errors.New("not unique short url")
	errNowSuchUrl        = errors.New("unknown url")
)

type Postgres struct {
	pg string
}

func NewPostgres() *Postgres {
	return &Postgres{
		pg: "postrgres",
	}
}

func createConnection() *sql.DB {

	db, err := sql.Open("postgres", os.Getenv("PG_URL"))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

func (p *Postgres) SaveData(og, short string) (string, error) {
	if err := p.uniqueCheck(og, short); err != nil {
		return "", err
	}

	db := createConnection()
	defer db.Close()

	preparedStatement := `INSERT INTO urls (originalurl, shorturl) VALUES ($1, $2) RETURNING shorturl`

	var ret string

	err := db.QueryRow(preparedStatement, og, short).Scan(&ret)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func (p *Postgres) GetData(short string) (string, error) {
	db := createConnection()
	defer db.Close()

	preparedStatement := `SELECT originalurl FROM urls WHERE shorturl=$1`

	var ret string

	err := db.QueryRow(preparedStatement, short).Scan(&ret)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errNowSuchUrl
		}
		return "", err
	}

	return ret, nil
}

func (p *Postgres) uniqueCheck(og, short string) error {
	db := createConnection()
	defer db.Close()

	preparedStatement := `SELECT id FROM urls WHERE originalurl=$1`
	var id int

	if err := db.QueryRow(preparedStatement, og).Scan(&id); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if id != 0 {
		return errNotUnique
	}

	preparedStatement = `SELECT id FROM urls WHERE shorturl=$1`

	if err := db.QueryRow(preparedStatement, short).Scan(&id); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if id != 0 {
		return errNotUniqueShortUrl
	}

	return nil
}
