package bot

import (
	tele "gopkg.in/telebot.v3"
	"log"
)

func (b Bot) onStart(c tele.Context) error {
	if _, err := b.db.Users.ByID(c.Sender().ID); err != nil {
		if err := b.db.Users.Create(c.Sender().ID); err != nil {
			log.Println("New user: ", c.Sender())
			return err
		}
	}
	return c.Send(
		b.Text(c, "start", c.Sender()),
		b.Markup(c, "menu"),
	)
}
