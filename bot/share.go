package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
	"time"
)

func (b Bot) onShare(c tele.Context) error {
	if err := b.sendShareMessage(c, "share"); err != nil {
		return err
	}
	// TODO add a payload which writes into cache or xsync map
	return b.db.Users.SetState(c.Sender().ID, database.StateForwardMessage)
}

func (b Bot) onForwardMessage(c tele.Context) error {
	var (
		msg  = c.Message()
		user = middle.User(c)
	)

	b.Delete(user.ShareMessage())

	if !msg.IsForwarded() {
		return b.sendShareMessage(c, "not_forward")
	}

	var (
		sender    = c.Sender()
		from      = sender.ID
		orgSender = msg.OriginalSender
		forward   = orgSender.ID
	)

	if _, err := b.db.Users.ByID(forward); err != nil {
		if _, err := b.db.ShareList.ByID(from, forward); err == nil {
			return b.sendShareMessage(c, "shared")
		}
		return b.sendShareMessage(c, "user_not_exist")
	}

	info := database.ShareList{
		FromUser:     from,
		FromUserName: sender.FirstName,
		ForwardFrom:  forward,
		CreatedAt:    time.Now(),
	}

	if err := b.db.ShareList.Add(info); err != nil {
		return err
	}

	_, err := b.Send(
		orgSender,
		b.Text(c, "share_info", info),
	)
	if err != nil {
		return err
	}

	b.db.Users.SetState(from, database.StateIdle)
	return c.Send(b.Text(c, "success_share", orgSender))
}

func (b Bot) sendShareMessage(c tele.Context, key string) error {
	msgShare, err := b.Send(
		c.Sender(),
		b.Text(c, key),
	)
	if err != nil {
		return err
	}

	user := middle.User(c)
	user.UpdateCache("ShareMessageID", msgShare.ID)
	return b.db.Users.SetCache(user)
}
