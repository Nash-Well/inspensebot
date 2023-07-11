package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	ShareListStorage interface {
		Add(s ShareList) error
		ByID(from int64, forward int64) (ShareList, error)
		ForwardList(userID int64) ([]ShareList, error)
		FromList(userID int64) ([]ShareList, error)
		DeleteFromList(userID int64) error
	}

	ShareLists struct {
		*sqlx.DB
	}

	ShareList struct {
		ID              int       `db:"id"`
		FromUser        int64     `db:"from_user"`
		FromUserName    string    `db:"from_name"`
		ForwardFrom     int64     `db:"forward_from"`
		ForwardUserName string    `db:"forward_name"`
		ShareType       string    `db:"share_type"`
		CreatedAt       time.Time `db:"created_at"`
	}

	ViewFinance struct {
		Finance
		Page      int
		ShareType string
	}
)

func (db *ShareLists) Add(s ShareList) error {
	const q = `INSERT INTO share_list(from_user, forward_from, from_name, share_type) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(q, s.FromUser, s.ForwardFrom, s.FromUserName, s.ShareType)
	return err
}

func (db *ShareLists) ByID(from int64, forward int64) (s ShareList, _ error) {
	const q = `SELECT * FROM share_list WHERE from_user=$1 AND forward_from=$2`
	return s, db.Get(&s, q, from, forward)
}

func (db *ShareLists) ForwardList(userID int64) (s []ShareList, _ error) {
	const q = `SELECT * FROM share_list WHERE forward_from=$1`
	return s, db.Select(&s, q, userID)
}

func (db *ShareLists) FromList(userID int64) (s []ShareList, _ error) {
	const q = `SELECT * FROM share_list WHERE from_user=$1`
	return s, db.Select(&s, q, userID)
}

func (db *ShareLists) DeleteFromList(userID int64) error {
	const q = `DELETE FROM share_list WHERE from_user=$1`
	_, err := db.Exec(q, userID)
	return err
}
