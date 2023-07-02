package middle

import (
	"database/sql"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
	"strings"
)

func SetUser(db *database.DB) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if strings.HasPrefix(c.Text(), "/start") {
				return next(c)
			}

			id := c.Sender().ID
			user, err := db.Users.ByID(id)
			if err == sql.ErrNoRows {
				if err := db.Users.Create(id); err != nil {
					return err
				}

				user, err = db.Users.ByID(id)
				if err != nil {
					return err
				}
			}

			c.Set("user", &user)
			return next(c)
		}
	}
}

func User(c tele.Context) *database.User {
	user, _ := c.Get("user").(*database.User)
	return user
}
