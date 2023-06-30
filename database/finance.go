package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	FinanceStorage interface {
		Create(f Finance) error
		CategoryCount(userID int64) (int, error)
		CategoryList(userID int64, page int) ([]Finance, error)
		//List(userID int64) ([]Finance, error)
		//ByID(id int) (Finance, error)
	}

	Finances struct {
		*sqlx.DB
	}

	Finance struct {
		ID          int       `db:"id"`
		UserID      int64     `db:"user_id,omitempty"`
		Type        string    `db:"type,omitempty"`
		Date        time.Time `db:"date,omitempty"`
		Amount      float64   `db:"amount,omitempty"`
		Category    string    `db:"category,omitempty"`
		Subcategory string    `db:"subcategory,omitempty"`
	}
)

func (db *Finances) Create(f Finance) error {
	const q = `INSERT INTO finances(user_id, type, date, amount, category, subcategory) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(q, f.UserID, f.Type, f.Date, f.Amount, f.Category, f.Subcategory)
	return err
}

func (db *Finances) CategoryList(userID int64, page int) (f []Finance, err error) {
	const q = `SELECT category FROM finances WHERE user_id=$1 GROUP BY category ORDER BY category DESC LIMIT 4 OFFSET $2`
	offset := page * 4
	return f, db.Select(&f, q, userID, offset)
}

func (db *Finances) CategoryCount(userID int64) (c int, _ error) {
	const q = `SELECT COUNT(*) FROM finances WHERE user_id=$1 GROUP BY category ORDER BY category DESC`
	return c, db.Get(&c, q, userID)
}
