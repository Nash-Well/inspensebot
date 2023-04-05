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

	defer finCache.Delete(userID)

	finance.Date = t
	finCache.Store(userID, finance)

	return c.Send(b.Text(c, "fin_added"))
}

func userCache(userID int64) (database.Finance, error) {
	finance, ok := finCache.Load(userID)
	if !ok {
		return database.Finance{}, errors.New("no such finance in cache")
	}

	return finance, nil
}
