package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
)

func (b Bot) onReport(c tele.Context) error {
	var (
		sender = c.Sender()
		markup = b.listMarkup(c)
		ml     = len(markup.InlineKeyboard) - 1
	)

	count, err := b.db.Finances.ListCount(sender.ID)
	if count == 0 {
		return c.Send(b.Text(c, "list_no_finances"))
	}

	markup.InlineKeyboard = markup.InlineKeyboard[ml-1 : ml]

	msgList, err := b.Send(
		sender,
		b.Text(c, "search_type"),
		markup,
	)
	if err != nil {
		return err
	}

	user := middle.User(c)
	user.UpdateCache("ListMessageID", msgList.ID)
	return b.db.Users.SetCache(user)
}
