package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"inspense-bot/bot/types"
	"time"
)

type (
	FinanceStorage interface {
		ByID(id int) (Finance, error)
		Create(f Finance) (int, error)
		UpdateFinance(id int, f Finance) error
		CategoryCount(f Finance) (int, error)
		CategoryList(u *User, f Finance) ([]string, error)
		UserByOffset(u *User) (Finance, error)
		ListCount(userID int64) (int, error)
		FinanceByOffset(vf ViewFinance) (Finance, error)
		ViewCount(userID int64, shareType string) (int, error)
		SearchByOffset(user *User, search types.Search) (Finance, error)
		SearchCount(userID int64, search types.Search) (count int, _ error)
		FinanceDetailed(search types.Search, userID int64) ([]int, error)
		Finances(userID int64, search types.Search) (map[string][]Finance, error)
		DeleteFinance(id int) error
		SearchByDate(fd string) ([]Finance, error)
		SearchByTypeDate(ft, fd string) ([]Finance, error)
		SearchByAll(ft, fd, fc string) ([]Finance, error)
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
		Location    []byte    `db:"location,omitempty"`
	}
)

func (db *Finances) ByID(id int) (f Finance, _ error) {
	const q = `SELECT * FROM finances WHERE id=$1`
	return f, db.Get(&f, q, id)
}

func (db *Finances) Create(f Finance) (int, error) {
	const q = `INSERT INTO finances(user_id, type, date, amount, category, subcategory, location) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var id int
	err := db.QueryRow(q, f.UserID, f.Type, f.Date, f.Amount, f.Category, f.Subcategory, f.Location).Scan(&id)
	return id, err
}

func (db *Finances) UpdateFinance(id int, f Finance) error {
	data := map[string]any{
		"user_id":     f.UserID,
		"type":        f.Type,
		"date":        f.Date,
		"amount":      f.Amount,
		"category":    f.Category,
		"subcategory": f.Subcategory,
	}

	query, args, err := squirrel.
		Update("finances").
		SetMap(data).
		Where(squirrel.Eq{"id": id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
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
	const q = `
		SELECT DISTINCT ON (user_id, type, category, subcategory, date, amount)
		MAX(id) AS id, user_id, type, MAX(date) AS date, MAX(amount) AS amount, MAX(category) AS category,
		MAX(subcategory) AS subcategory
		FROM finances
		WHERE user_id = $1
		GROUP BY user_id, type, category, subcategory, date, amount
		ORDER BY user_id, type, category, subcategory, date DESC, amount DESC, id DESC
		LIMIT 1 OFFSET $2;`
	return f, db.Get(&f, q, u.ID, u.GetCache().ListPage)
}

func (db *Finances) ListCount(userID int64) (c int, _ error) {
	const q = `SELECT COUNT(*) FROM (
			SELECT DISTINCT ON (user_id, type, category, subcategory, date, amount)
				MAX(id) AS id, user_id, type, MAX(date) AS date, MAX(amount) AS amount, MAX(category) AS category,
				MAX(subcategory) AS subcategory
			FROM finances
			WHERE user_id = $1
			GROUP BY user_id, type, category, subcategory, date, amount
		) AS count;`
	return c, db.Get(&c, q, userID)
}

func (db *Finances) FinanceByOffset(vf ViewFinance) (f Finance, _ error) {
	const q = `
		SELECT DISTINCT ON (user_id, type, category, subcategory, date, amount)
		id, user_id, type, date, amount, category, subcategory
		FROM finances
		WHERE user_id = $1
			AND (
				($2 = '' AND type IN ('income', 'expense'))
				OR type = $2
			)
		ORDER BY user_id, type, category, subcategory, date DESC, amount DESC, id DESC
		LIMIT 1 OFFSET $3;`
	return f, db.Get(&f, q, vf.UserID, vf.ShareType, vf.Page)
}

func (db *Finances) ViewCount(userID int64, shareType string) (c int, _ error) {
	const q = `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT ON (user_id, type, category, subcategory, date, amount)
				id, user_id, type, date, amount, category, subcategory
			FROM finances
			WHERE user_id = $1
				AND (
					($2 = '' AND type IN ('income', 'expense'))
					OR type = $2
				)
			ORDER BY user_id, type, category, subcategory, date DESC, amount DESC, id DESC
		) AS count;`
	return c, db.Get(&c, q, userID, shareType)
}

func (db *Finances) FinanceDetailed(search types.Search, userID int64) (res []int, _ error) {
	const q = `
		SELECT
  			CASE
    			WHEN $1 = 'year' THEN EXTRACT(YEAR FROM date)
    			WHEN $1 = 'month' THEN EXTRACT(MONTH FROM date)
    			WHEN $1 = 'day' THEN EXTRACT(DAY FROM date)
  			END AS result
		FROM finances WHERE user_id = $3 AND (
    		($1 = 'year')
   			OR ($1 = 'month' AND EXTRACT(YEAR FROM date) = $4)
    		OR ($1 = 'day' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5)
  		) AND type = $2
		GROUP BY result ORDER BY result DESC;`
	return res, db.Select(&res, q, search.Search, search.Type, userID, search.Year, search.Month)
}

func (db *Finances) SearchCount(userID int64, search types.Search) (count int, _ error) {
	const q = `SELECT COUNT(*) AS record_count
			   FROM finances WHERE user_id = $3 AND (
    		   (
    		   	 $1 = ''
    			 OR $1 = 'year' AND EXTRACT(YEAR FROM date) = $4)
    			 OR ($1 = 'month' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5)
    			 OR ($1 = 'day' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5 AND EXTRACT(DAY FROM date) = $6)
  			   ) AND type = $2;
`
	return count, db.Get(&count, q, search.Search, search.Type, userID, search.Year, search.Month, search.Day)
}

func (db *Finances) SearchByOffset(user *User, search types.Search) (f Finance, _ error) {
	const q = `SELECT * FROM finances WHERE user_id = $3 AND (
    		   (
    		     $1 = ''
    		     OR $1 = 'year' AND EXTRACT(YEAR FROM date) = $4)
    			 OR ($1 = 'month' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5)
    			 OR ($1 = 'day' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5 AND EXTRACT(DAY FROM date) = $6)
  			   ) AND type = $2 LIMIT 1 OFFSET $7`
	return f, db.Get(&f, q, search.Search, search.Type, user.ID, search.Year, search.Month, search.Day, user.GetCache().ListPage)
}

func (db *Finances) Finances(userID int64, search types.Search) (map[string][]Finance, error) {
	const q = `
		SELECT date, sum(amount) as amount, category
		FROM finances
		WHERE user_id = $3 AND (
			($1 = '' 
		    OR ($1 = 'year' AND EXTRACT(YEAR FROM date) = $4))
			OR ($1 = 'month' AND EXTRACT(YEAR FROM date) = $4 AND EXTRACT(MONTH FROM date) = $5)
		) AND type = $2 GROUP BY category, date
		ORDER BY EXTRACT(MONTH FROM date), category ASC`

	rows, err := db.Query(q, search.Search, search.Type, userID, search.Year, search.Month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	financeMap := make(map[string][]Finance)

	for rows.Next() {
		var (
			date     time.Time
			amount   float64
			category string
		)

		err := rows.Scan(&date, &amount, &category)
		if err != nil {
			return nil, err
		}

		dateStr := date.Format("01")
		if search.Search == "month" {
			dateStr = date.Format("02")
		}

		finance := Finance{
			Amount:   amount,
			Category: category,
			Date:     date,
		}

		if _, ok := financeMap[dateStr]; !ok {
			financeMap[dateStr] = []Finance{}
		}
		financeMap[dateStr] = append(financeMap[dateStr], finance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return financeMap, nil
}

func (db *Finances) DeleteFinance(id int) error {
	const q = "DELETE FROM finances WHERE id=$1"
	_, err := db.Exec(q, id)
	return err
}

func (db *Finances) SearchByDate(fd string) (f []Finance, _ error) {
	const q = "SELECT * FROM finances WHERE date=$1"
	return f, db.Select(&f, q, fd)
}

func (db *Finances) SearchByTypeDate(ft, fd string) (f []Finance, _ error) {
	const q = "SELECT * FROM finances WHERE type=$1 AND date=$2"
	return f, db.Select(&f, q, ft, fd)
}

func (db *Finances) SearchByAll(ft, fd, fc string) (f []Finance, _ error) {
	const q = "SELECT * FROM finances WHERE type=$1 AND date=$2 AND category ilike '%' || $3 || '%'"
	return f, db.Select(&f, q, ft, fd, fc)
}
