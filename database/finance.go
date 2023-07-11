package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	FinanceStorage interface {
		Create(f Finance) (int, error)
		CategoryCount(f Finance) (int, error)
		CategoryList(u *User, f Finance) ([]string, error)
		UserByOffset(u *User) (Finance, error)
		ListCount(userID int64) (int, error)
		FinanceByOffset(vf ViewFinance) (Finance, error)
		ViewCount(userID int64, shareType string) (c int, _ error)
		//ByID(id int) (Finance, error)
	}

	Finances struct {
		*sqlx.DB
	}

	Finance struct {
		ID          int       `db:"id, omitempty"`
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

func (db *Finances) CategoryList(u *User, f Finance) (c []string, err error) {
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

func (db *Finances) UserByOffset(u *User) (f Finance, _ error) {
	const q = `SELECT MAX(id) as id, user_id, type, MAX(date) as date, MAX(amount) as amount, MAX(category) as category,
               MAX(subcategory) as subcategory FROM finances WHERE user_id=$1 GROUP BY user_id, type, category, 
               subcategory ORDER BY type, date,id LIMIT 1 OFFSET $2`
	return f, db.Get(&f, q, u.ID, u.GetCache().ListPage)
}

func (db *Finances) ListCount(userID int64) (c int, _ error) {
	const q = `SELECT count(*) FROM (SELECT MAX(id) as id, user_id, type, MAX(date) as date, MAX(amount) as amount, 
			   MAX(category) as category, MAX(subcategory) as subcategory FROM finances WHERE user_id=$1 
			   GROUP BY user_id, type, category, subcategory) as count`
	return c, db.Get(&c, q, userID)
}

func (db *Finances) FinanceByOffset(vf ViewFinance) (f Finance, _ error) {
	const q = `
		SELECT MAX(id) AS id, user_id, type, MAX(date) AS date, MAX(amount) AS amount, MAX(category) AS category,
		MAX(subcategory) AS subcategory	FROM finances WHERE user_id = $1
		AND (
   		($2 = '' AND type IN ('income', 'expense'))
   		OR type = $2
   		)
		GROUP BY user_id, type, category, subcategory
		ORDER BY type, date, id
		LIMIT 1 OFFSET $3;
   `
	return f, db.Get(&f, q, vf.UserID, vf.ShareType, vf.Page)
}

func (db *Finances) ViewCount(userID int64, shareType string) (c int, _ error) {
	const q = `
		SELECT COUNT(*) FROM
		(SELECT MAX(id) AS id, user_id, type, MAX(date) AS date, MAX(amount) AS amount, MAX(category) AS category, 
		MAX(subcategory) AS subcategory	FROM finances WHERE user_id = $1 
		AND (
   		($2 = '' AND type IN ('income', 'expense'))
   		OR type = $2
   		)
		GROUP BY user_id, type, category, subcategory
		ORDER BY type, date, id) as count
	`
	return c, db.Get(&c, q, userID, shareType)
}
