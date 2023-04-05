package bot

import (
	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
)

var finCache = xsync.NewIntegerMapOf[int64, database.Finance]()

func (b Bot) onAdd(c tele.Context) error {
	if err := b.db.Users.SetState(c.Sender().ID, database.AddingTypeState); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "add_fin"),
		b.Markup(c, "type_menu"),
	)
}

func (b Bot) onType(c tele.Context) error {
	id := c.Sender().ID

	finCache.Store(
		id,
		database.Finance{Type: c.Data()},
	)

	if err := b.db.Users.SetState(id, database.AddingAmount); err != nil {
		return err
	}

	return c.EditOrSend(b.Text(c, "add_amount"))
}
