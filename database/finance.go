package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	FinanceStorage interface {
		Create(userID int64, f Finance) error
		List(userID int64) ([]Finance, error)
		ByID(id int) (Finance, error)
	}

	Finances struct {
		*sqlx.DB
	}

	Finance struct {
		ID       int       `db:"id"`
		UserID   int64     `db:"user_id,omitempty"`
		Type     string    `db:"type,omitempty"`
		Date     time.Time `db:"date,omitempty"`
		Amount   float64   `db:"amount,omitempty"`
		Category string    `db:"category,omitempty"`
	}
)

func (db *Finances) ByID(id int) (f Finance, _ error) {
	const q = "SELECT * FROM finances WHERE id=$1"
	return f, db.Get(&f, q, id)
}

func (db *Finances) Create(userID int64, f Finance) error {
	const q = `INSERT INTO finances(user_id, type, date, amount, category) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(q, userID, f.Type, f.Date, f.Amount, f.Category)
	return err
}

func (db *Finances) List(userID int64) (f []Finance, _ error) {
	const q = "SELECT * FROM finances WHERE user_id=$1"
	return f, db.Select(&f, q, userID)
}
