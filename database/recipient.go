package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	RecipientStorage interface {
		Add(r Recipient) error
		ByID(finID int) (Recipient, error)
	}

	Recipients struct {
		*sqlx.DB
	}

	Recipient struct {
		ID           int       `db:"id"`
		FinanceID    int       `db:"finance_id"`
		Media        string    `db:"media"`
		MediaType    string    `db:"media_type"`
		MediaCaption string    `db:"media_caption"`
		CreatedAt    time.Time `db:"created_at"`
	}
)

func (db *Recipients) Add(r Recipient) error {
	const q = `INSERT INTO recipient (finance_id, media, media_type, media_caption) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(q, r.FinanceID, r.Media, r.MediaType, r.MediaCaption)
	return err
}

func (db *Recipients) ByID(finID int) (r Recipient, _ error) {
	const q = `SELECT * FROM recipient WHERE finance_id=$1`
	return r, db.Get(&r, q, finID)
}
