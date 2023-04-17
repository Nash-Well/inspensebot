package bot

import (
	"errors"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
)

func (b Bot) onText(c tele.Context) error {
	state, err := b.db.Users.State(c.Sender().ID)
	if err != nil {
		return err
	}

	switch state {
	case database.AddingAmount:
		return b.onAmount(c)
	case database.AddingDate:
		return b.onDate(c)
	default:
		return nil
	}

}

func (b Bot) onAmount(c tele.Context) error {
	var (
		userID = c.Sender().ID
		msg    = c.Text()
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(msg, 64)
	if err != nil {
		return err
	}

	finance.UserID = userID
	finance.Amount = amount

	finCache.Store(userID, finance)

	if err := b.db.Users.SetState(userID, database.AddingDate); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "add_date"),
	)
}

func (b Bot) onDate(c tele.Context) error {
	var (
		msg    = c.Text()
		userID = c.Sender().ID
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	t, err := time.Parse("02.01.2006", msg)
	if err != nil {
		return c.Send(b.Text(c, "error_date"))
	}

	finance.Date = t
	finCache.Store(userID, finance)

	if err := b.db.Users.SetState(userID, database.AddingMedia); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "recipient"),
		b.Markup(c, "recipient_menu"),
	)
}

func userCache(userID int64) (database.Finance, error) {
	finance, ok := finCache.Load(userID)
	if !ok {
		return database.Finance{}, errors.New("no such finance in cache")
	}

	return finance, nil
}

func (b Bot) onMedia(c tele.Context) error {
	var (
		mediaType = c.Message().Media().MediaType()
		mediaID   = c.Message().Media().MediaFile().FileID
	)

	defer b.db.Users.SetState(c.Sender().ID, database.DefaultState)

	switch mediaType {
	case "photo":
		_, err := b.Send(
			c.Chat(),
			&tele.Photo{File: tele.File{FileID: mediaID}},
		)
		return err
	case "document":
		_, err := b.Send(
			c.Chat(),
			&tele.Document{File: tele.File{FileID: mediaID}},
		)
		return err
	default:
		return nil
	}
}
