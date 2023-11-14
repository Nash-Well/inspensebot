package database

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"time"
)

type (
	RecipientStorage interface {
		Add(r Recipient) error
		ByID(finID int) (Recipient, error)
		DeleteRecipient(id int) error
		UpdateRecipient(r Recipient) error
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

func (db *Recipients) DeleteRecipient(id int) error {
	const q = "DELETE FROM recipient WHERE finance_id=$1"
	_, err := db.Exec(q, id)
	return err
}

func (db *Recipients) UpdateRecipient(r Recipient) error {
	data := map[string]any{
		"media":         r.Media,
		"media_type":    r.MediaType,
		"media_caption": r.MediaCaption,
	}

	query, args, err := squirrel.
		Update("recipient").
		SetMap(data).
		Where(squirrel.Eq{"finance_id": r.FinanceID}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}
