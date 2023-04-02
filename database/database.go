package database

import (
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose"
)

type DB struct {
	*sqlx.DB
	Users UserStorage
}

func Open(url string) (*DB, error) {
	db, err := sqlx.Connect("pgx", url)
	if err != nil {
		return nil, err
	}

	return &DB{
		DB:    db,
		Users: &Users{DB: db},
	}, nil
}

func (db *DB) Migrate() error {
	return goose.Up(db.DB.DB, "sql")
}
