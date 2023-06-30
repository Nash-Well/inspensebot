package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
)

func (b Bot) onText(c tele.Context) error {
	state, err := b.db.Users.State(c.Sender().ID)
	if err != nil {
		return err
	}

	defer b.deleteWithReply(c)

	switch state {
	case database.StateAddAmount:
		return b.onAmount(c)
	case database.StateAddDate:
		return b.onDate(c)
	case database.StateAddCategory:
		return b.onCategory(c)
	case database.StateAddSubCategory:
		return b.onSubCategory(c)
	default:
		return nil
	}

}

func (b Bot) deleteWithReply(c tele.Context) error {
	if err := c.Delete(); err != nil {
		return err
	}
	if reply := c.Message().ReplyTo; reply != nil {
		return b.Delete(reply)
	}
	return nil
}

func (b Bot) onMedia(c tele.Context) error {
	var (
		mediaType = c.Message().Media().MediaType()
		mediaID   = c.Message().Media().MediaFile().FileID
	)

	defer b.db.Users.SetState(c.Sender().ID, database.StateIdle)

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
