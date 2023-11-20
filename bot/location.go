package bot

import (
	"encoding/json"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
)

func (b Bot) onLocation(c tele.Context) error {
	user := middle.User(c)

	defer b.deleteWithReply(c)

	switch user.State {
	case database.StateAddLocation:
		return b.onAddUserLocation(c)
	default:
		return nil
	}
}

func (b Bot) onAddUserLocation(c tele.Context) error {
	var (
		msg    = c.Message()
		loc    = msg.Location
		sender = c.Sender()
	)

	ml, err := json.Marshal(loc)
	if err != nil {
		return err
	}

	uc, err := userCache(sender.ID)
	if err != nil {
		return err
	}

	uc.Location = ml
	finCache.Store(sender.ID, uc)

	if err := c.EditOrSend(
		b.Text(c, "recipient_exists"),
		b.Markup(c, "recipient_menu"),
	); err != nil {
		return err
	}

	user := middle.User(c)

	if user.Exists("CategoryMessageID") {
		b.Delete(user.CategoryMessage())
		user.DeleteFromCache("CategoryMessageID")
		b.db.Users.SetCache(user)
	}

	return b.db.Users.SetState(sender.ID, database.StateEditRecipient)
}

func (b Bot) onBackToListActions(c tele.Context) error {
	user := middle.User(c)

	b.Delete(user.ListMessage())
	b.Delete(user.ListActionsMessage())

	return b.onFunctions(c)
}

func (b Bot) onLocationChoice(c tele.Context) error {
	userID := c.Sender().ID

	switch c.Data() {
	case "loc_agr":
		c.Delete()

		msgLoc, err := b.Send(
			c.Sender(),
			b.Text(c, "add_location"),
			b.Markup(c, "user_location"),
			tele.ForceReply,
		)
		if err != nil {
			return err
		}

		user := middle.User(c)
		user.UpdateCache("CategoryMessageID", msgLoc.ID)
		b.db.Users.SetCache(user)

		return b.db.Users.SetState(c.Sender().ID, database.StateAddLocation)
	case "loc_not":
		defer b.db.Users.SetState(userID, database.StateEditRecipient)

		return c.EditOrSend(
			b.Text(c, "recipient_exists"),
			b.Markup(c, "recipient_menu"),
		)
	default:
		return nil
	}
}
