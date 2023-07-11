package bot

import (
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
	"strconv"
	"strings"
	"time"
)

type ViewFinance struct {
	database.Finance
	Page      int
	ShareType string
}

func (b Bot) onShare(c tele.Context) error {
	if err := b.sendShareMessage(c, "share"); err != nil {
		return err
	}

	user := middle.User(c)
	user.UpdateCache("PayloadType", c.Message().Payload)
	b.db.Users.SetCache(user)

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

	if from == forward {
		return b.sendShareMessage(c, "same_forward")
	}

	if _, err := b.db.Users.ByID(forward); err != nil {
		return b.sendShareMessage(c, "user_not_exist")
	}

	if _, err := b.db.ShareList.ByID(from, forward); err == nil {
		return b.sendShareMessage(c, "shared")
	}

	info := database.ShareList{
		FromUser:     from,
		FromUserName: sender.FirstName,
		ForwardFrom:  forward,
		CreatedAt:    time.Now(),
		ShareType:    user.GetCache().PayloadType,
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

func (b Bot) onView(c tele.Context) error {
	list, err := b.db.ShareList.FromList(c.Sender().ID)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return c.Send(b.Text(c, "empty_list"))
	}

	var row tele.Row
	for _, v := range list {
		row = append(row, *b.Button(c, "from_user", v))
	}

	markup := b.NewMarkup()
	markup.Inline(markup.Split(1, row)...)

	msgView, err := b.Send(
		c.Sender(),
		b.Text(c, "view", list),
		markup,
	)
	if err != nil {
		return err
	}

	user := middle.User(c)
	user.UpdateCache("ViewMessageID", msgView.ID)
	return b.db.Users.SetCache(user)
}

func (b Bot) onUser(c tele.Context) error {
	var (
		args      = c.Args()
		userID, _ = strconv.ParseInt(args[0], 10, 64)
	)

	finance, err := b.financeViewExt(userID, args[1], 0)
	if err != nil {
		return err
	}

	return b.constructView(c, finance)
}

func (b Bot) onBackView(c tele.Context) error {
	var (
		args      = c.Args()
		page, _   = strconv.Atoi(args[0])
		userID, _ = strconv.ParseInt(args[1], 10, 64)
		shareType = args[2]
	)

	page -= 1
	if page < 0 {
		count, err := b.db.Finances.ViewCount(userID, shareType)
		if err != nil {
			return err
		}

		if err := b.ShowAlert(c, count); err != nil {
			return err
		}

		if count > 1 {
			page = count - 1
		}
	}

	finance, err := b.financeViewExt(userID, shareType, page)
	if err != nil {
		return err
	}

	return b.constructView(c, finance)
}

func (b Bot) onForwardView(c tele.Context) error {
	var (
		args      = c.Args()
		page, _   = strconv.Atoi(args[0])
		userID, _ = strconv.ParseInt(args[1], 10, 64)
		shareType = args[2]
	)

	page += 1
	count, err := b.db.Finances.ViewCount(userID, shareType)
	if err != nil {
		return err
	}

	if err := b.ShowAlert(c, count); err != nil {
		return err
	}

	if count > 1 && page > count-1 {
		page = 0
	}

	finance, err := b.financeViewExt(userID, args[2], page)
	if err != nil {
		return err
	}

	return b.constructView(c, finance)
}

func (b Bot) constructView(c tele.Context, finance ViewFinance) (err error) {
	_, err = b.Edit(
		middle.User(c).ViewMessage(),
		b.Text(c, "list", finance),
		b.Markup(c, "view_menu", finance),
	)

	return
}

func (b Bot) financeViewExt(userID int64, shareType string, page int) (ViewFinance, error) {
	finance, err := b.db.Finances.FinanceByOffset(userID, shareType, page)
	if err != nil {
		return ViewFinance{}, err
	}
	finance.Type = strings.Title(finance.Type)

	f_ext := ViewFinance{
		Finance:   finance,
		Page:      page,
		ShareType: shareType,
	}

	return f_ext, nil
}
