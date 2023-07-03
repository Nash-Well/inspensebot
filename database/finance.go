package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	FinanceStorage interface {
		Create(f Finance) (int, error)
		CategoryCount(f Finance) (int, error)
		CategoryList(u User, f Finance) ([]string, error)
		//List(userID int64) ([]Finance, error)
		//ByID(id int) (Finance, error)
	}

	Finances struct {
		*sqlx.DB
	}

	Finance struct {
		UserID      int64     `db:"user_id,omitempty"`
		Type        string    `db:"type,omitempty"`
		Date        time.Time `db:"date,omitempty"`
		Amount      float64   `db:"amount,omitempty"`
		Category    string    `db:"category,omitempty"`
		Subcategory string    `db:"subcategory,omitempty"`
	}
)

func (db *Finances) Create(f Finance) (int, error) {
	const q = `INSERT INTO finances(user_id, type, date, amount, category, subcategory) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int
	err := db.QueryRow(q, f.UserID, f.Type, f.Date, f.Amount, f.Category, f.Subcategory).Scan(&id)
	return id, err
}

func (db *Finances) CategoryList(u User, f Finance) (c []string, err error) {
	const q = `SELECT category FROM finances WHERE user_id=$1 AND type=$2 GROUP BY category ORDER BY category DESC LIMIT 4 OFFSET $3`
	page := u.GetCache().CategoryPage
	offset := page * 4
	return c, db.Select(&c, q, u.ID, f.Type, offset)
}

func (db *Finances) CategoryCount(f Finance) (c int, _ error) {
	const q = `SELECT COUNT(*) FROM (
    				SELECT category FROM finances WHERE user_id = $1 AND type = $2 GROUP BY category
				) AS categories`
	return c, db.Get(&c, q, f.UserID, f.Type)
}
