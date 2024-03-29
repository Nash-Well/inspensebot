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
	case database.StateForwardMessage:
		return b.onForwardMessage(c)
	case database.StateEditAmount:
		return b.onEditedAmount(c)
	case database.StateEditCategory:
		return b.onEditedCategory(c)
	case database.StateEditSubCategory:
		return b.onEditedSubcategory(c)
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
	state, err := b.db.Users.State(c.Sender().ID)
	if err != nil {
		return nil
	}

	switch state {
	case database.StateAddMedia:
		return b.onAddMedia(c)
	case database.StateEditRecipient:
		return b.onEditedRecipient(c)
	case database.StateForwardMessage:
		return b.onForwardMessage(c)
	default:
		return nil
	}
}
