package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
	"log"
)

func (b Bot) onRecipient(c tele.Context) error {
	var (
		data   = c.Data()
		userID = c.Sender().ID
	)

	switch data {
	case "exists":
		state, _ := b.db.Users.State(userID)
		log.Println(state)
		return c.EditOrSend(
			b.Text(c, "recipient"),
			b.Markup(c, "cancel_opts"),
		)
	case "not_exists":
		defer b.db.Users.SetState(userID, database.DefaultState)
		defer finCache.Delete(userID)

		if err := c.Delete(); err != nil {
			return err
		}
		return c.Send(b.Text(c, "fin_added"))
	default:
		return nil
	}

}

func (b Bot) onCancel(c tele.Context) error {
	return c.EditOrSend(
		b.Text(c, "recipient"),
		b.Markup(c, "recipient_menu"),
	)
}
