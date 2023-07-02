package bot

import (
	"database/sql"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
)

func (b Bot) onStart(c tele.Context) error {
	userID := c.Sender().ID
	if middle.User(c) == nil {
		_, err := b.db.Users.ByID(userID)
		if err == sql.ErrNoRows {
			if err := b.db.Users.Create(userID); err != nil {
				return err
			}
		}
	}

	return c.Send(
		b.Text(c, "start", c.Sender()),
		b.Markup(c, "menu"),
	)
}
