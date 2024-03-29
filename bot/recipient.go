package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
)

func (b Bot) onRecipient(c tele.Context) error {
	userID := c.Sender().ID

	switch c.Data() {
	case "exists":
		b.db.Users.SetState(userID, database.StateAddMedia)

		c.Delete()

		msgMedia, err := b.Send(
			c.Sender(),
			b.Text(c, "recipient"),
			b.Markup(c, "cancel_opts"),
		)
		if err != nil {
			return err
		}

		user := middle.User(c)
		user.UpdateCache("MediaMessageID", msgMedia.ID)
		return b.db.Users.SetCache(user)
	case "not_exists":
		defer b.db.Users.SetState(userID, database.StateIdle)
		c.Delete()

		finance, err := userCache(userID)
		if err != nil {
			return err
		}
		defer finCache.Delete(userID)

		return b.addFinance(c, finance)
	default:
		return nil
	}
}

func (b Bot) onCancel(c tele.Context) error {
	_, err := b.Edit(
		middle.User(c).MediaMessage(),
		b.Text(c, "recipient_exists"),
		b.Markup(c, "recipient_menu"),
	)

	return err
}

func (b Bot) onAddMedia(c tele.Context) error {
	var (
		text      = c.Text()
		userID    = c.Sender().ID
		media     = c.Message().Media()
		mediaID   = media.MediaFile().FileID
		mediaType = media.MediaType()
	)

	c.Delete()
	b.Delete(middle.User(c).MediaMessage())

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	id, err := b.db.Finances.Create(finance)
	if err != nil {
		return err
	}
	defer finCache.Delete(userID)

	r := database.Recipient{
		FinanceID:    id,
		Media:        mediaID,
		MediaType:    mediaType,
		MediaCaption: text,
	}

	if err := b.db.Recipients.Add(r); err != nil {
		return err
	}

	b.db.Users.SetState(userID, database.StateIdle)

	return c.Send(
		b.Text(c, "fin_added"),
		b.Markup(c, "menu"),
	)
}
