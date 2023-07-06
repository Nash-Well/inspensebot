package bot

import (
	"database/sql"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
	"strconv"
	"strings"
)

type Finance struct {
	database.Finance
	database.Recipient
	Page int
}

func (b Bot) onList(c tele.Context) error {
	finance, err := b.financeExt(c, 0)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Send(
				b.Text(c, "list_no_finances"),
			)
		}
		return err
	}

	return b.constructList(c, finance)
}

func (b Bot) onBackList(c tele.Context) error {
	page, _ := strconv.Atoi(c.Data())

	page -= 1
	if page < 0 {
		count, err := b.db.Finances.ListCount(c.Sender().ID)
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

	finance, err := b.financeExt(c, page)
	if err != nil {
		return err
	}

	return b.constructList(c, finance, true)
}

func (b Bot) onForwardList(c tele.Context) error {
	page, _ := strconv.Atoi(c.Data())

	page += 1
	count, err := b.db.Finances.ListCount(c.Sender().ID)
	if err != nil {
		return err
	}

	if err := b.ShowAlert(c, count); err != nil {
		return err
	}

	if count > 1 && page > count-1 {
		page = 0
	}

	finance, err := b.financeExt(c, page)
	if err != nil {
		return err
	}

	return b.constructList(c, finance, true)
}

func (b Bot) onFunctions(c tele.Context) error {
	user := middle.User(c)

	finance, err := b.db.Finances.ByOffset(user)
	if err != nil {
		return err
	}
	finance.Type = strings.Title(finance.Type)

	r, err := b.db.Recipients.ByID(finance.ID)
	if err == nil && err == sql.ErrNoRows {
		r = database.Recipient{}
	}

	f_ext := Finance{
		Finance:   finance,
		Recipient: r,
	}

	b.Delete(user.ListMessage())
	return b.constructListActions(c, f_ext)
}

func (b Bot) onBackToList(c tele.Context) error {
	user := middle.User(c)

	finance, err := b.financeExt(c, user.GetCache().ListPage)
	if err != nil {
		return err
	}

	b.Delete(user.ListMessage())
	b.Delete(user.ListActionsMessage())

	return b.constructList(c, finance)
}

func (b Bot) ShowAlert(c tele.Context, count int) error {
	if count == 1 {
		return c.Respond(
			&tele.CallbackResponse{
				Text:      b.Text(c, "list_single_record"),
				ShowAlert: true,
			},
		)
	}

	return nil
}

func (b Bot) financeExt(c tele.Context, page int) (Finance, error) {
	user := middle.User(c)
	user.UpdateCache("ListPage", page)
	b.db.Users.SetCache(user)

	finance, err := b.db.Finances.ByOffset(user)
	if err != nil {
		return Finance{}, err
	}
	finance.Type = strings.Title(finance.Type)

	f_ext := Finance{
		Finance: finance,
		Page:    page,
	}

	return f_ext, nil
}

func (b Bot) constructList(c tele.Context, finance Finance, edit ...bool) (err error) {
	var (
		editing = len(edit) > 0 && edit[0]
		user    = middle.User(c)
		msgList *tele.Message
	)

	if editing {
		msgList = user.ListMessage()
	} else {
		msgList = c.Message()
	}

	if editing {
		msgList, err = b.Edit(
			msgList,
			b.Text(c, "list", finance),
			b.Markup(c, "list_menu", finance),
		)
	} else {
		msgList, err = b.Send(
			c.Sender(),
			b.Text(c, "list", finance),
			b.Markup(c, "list_menu", finance),
		)
	}

	user.UpdateCache("ListMessageID", msgList.ID)
	return b.db.Users.SetCache(user)
}

func (b Bot) constructListActions(c tele.Context, finance Finance, edit ...bool) (err error) {
	var (
		editing    = len(edit) > 0 && edit[0]
		user       = middle.User(c)
		msgList    *tele.Message
		msgActions *tele.Message
	)

	if editing {
		msgList = user.ListMessage()
	} else {
		msgList = c.Message()
	}

	var what any

	if finance.MediaType != "" && finance.Media != "" {
		what = b.Media(c, finance)
	} else {
		what = b.Text(c, "list", finance)
	}

	if editing {
		msgList, err = b.Edit(
			msgList,
			what,
		)
	} else {
		msgList, err = b.Send(
			c.Sender(),
			what,
		)
	}

	if editing {
		msgActions, err = b.Edit(
			msgActions,
			b.Text(c, "list_actions", finance),
			b.Markup(c, "list_opts", finance),
		)
	}
	msgActions, err = b.Send(
		c.Sender(),
		b.Text(c, "list_actions", finance),
		b.Markup(c, "list_opts", finance),
	)

	user.UpdateCache("ListMessageID", msgList.ID)
	user.UpdateCache("ActionsMessageID", msgActions.ID)
	return b.db.Users.SetCache(user)
}

func (b Bot) Media(c tele.Context, f Finance) tele.Media {
	switch f.MediaType {
	case "photo":
		return &tele.Photo{
			File:    tele.File{FileID: f.Media},
			Caption: b.Text(c, "list_ext", f),
		}
	case "document":
		return &tele.Document{
			File:    tele.File{FileID: f.Media},
			Caption: b.Text(c, "list_ext", f),
		}
	default:
		return nil
	}
}
