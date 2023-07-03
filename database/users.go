package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/telebot.v3"
	"time"
)

type State string

const (
	StateIdle           = ""
	StateAddType        = "adding_type"
	StateAddAmount      = "adding_amount"
	StateAddDate        = "adding_date"
	StateAddCategory    = "adding_category"
	StateAddSubCategory = "adding_subcategory"
	StateAddMedia       = "adding_media"
)

type (
	UserStorage interface {
		ByID(id int64) (User, error)
		Create(id int64) error
		SetState(id int64, state State) error
		State(id int64) (State, error)
		SetLanguage(id int64, lang string) error
		Language(chat Chat) (lang string, err error)
		SetCache(u User) error
	}

	Users struct {
		*sqlx.DB
	}

	Chat interface {
		telebot.Recipient
	}

	User struct {
		ID        int64          `db:"id,omitempty" json:"id"`
		CreatedAt time.Time      `db:"created_at,omitempty" json:"createdAt"`
		Language  string         `db:"language,omitempty" json:"language"`
		State     string         `db:"state,omitempty" json:"state"`
		Cache     types.JSONText `db:"cache" json:"cache"`
	}
)

func (db Users) ByID(id int64) (u User, _ error) {
	const q = "SELECT * FROM users WHERE id=$1"
	return u, db.Get(&u, q, id)
}

func (db Users) Create(id int64) error {
	const q = "INSERT INTO users(id) VALUES ($1)"
	_, err := db.Exec(q, id)
	return err
}

func (db *Users) SetState(id int64, state State) error {
	const q = "UPDATE users SET state=$1 WHERE id=$2"
	_, err := db.Exec(q, state, id)
	return err
}

func (db *Users) State(id int64) (s State, _ error) {
	const q = "SELECT state FROM users WHERE id=$1"
	return s, db.Get(&s, q, id)
}

func (db Users) SetLanguage(id int64, lang string) error {
	const q = "UPDATE users SET language=$1 WHERE id=$2"
	_, err := db.Exec(q, lang, id)
	return err
}

func (db Users) Language(chat Chat) (lang string, err error) {
	const q = "SELECT language FROM users WHERE id=$1"
	return lang, db.Get(&lang, q, chat.Recipient())
}

func (db *Users) SetCache(u User) error {
	const query = `UPDATE users SET cache=$1 WHERE id=$2`
	_, err := db.Exec(query, u.Cache, u.ID)
	return err
}
