package bot

import (
	"database/sql"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/database"
	"regexp"

	"strconv"
	"strings"
)

type Finance struct {
	database.Finance
	database.Recipient
	Page int
}

func (b Bot) onList(c tele.Context) error {
	var (
		user   = middle.User(c)
		sender = c.Sender()
	)

	count, err := b.db.Finances.ListCount(sender.ID)
	if count == 0 {
		return c.Send(b.Text(c, "list_no_finances"))
	}

	msgList, err := b.Send(
		sender,
		b.Text(c, "search_type"),
		b.listMarkup(c),
	)
	if err != nil {
		return err
	}

	user.UpdateCache("ListMessageID", msgList.ID)
	return b.db.Users.SetCache(user)
}

func (b Bot) onBackList(c tele.Context) error {
	var (
		page, _ = strconv.Atoi(c.Data())
		userID  = c.Sender().ID
	)

	page -= 1
	if page < 0 {
		count, err := b.db.Finances.ListCount(userID)
		if search, ok := searchPref.Load(userID); ok {
			count, err = b.db.Finances.SearchCount(userID, search)
		}

		if err != nil {
			return err
		}

		if count == 1 {
			return b.ShowAlert(c)
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
	var (
		page, _ = strconv.Atoi(c.Data())
		userID  = c.Sender().ID
	)

	page += 1
	count, err := b.db.Finances.ListCount(userID)

	if search, ok := searchPref.Load(userID); ok {
		count, err = b.db.Finances.SearchCount(userID, search)
	}
	if err != nil {
		return err
	}

	if count == 1 {
		return b.ShowAlert(c)
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

	finance, err := b.db.Finances.UserByOffset(user)
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

func (b Bot) onChangeType(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	if finance.Type == "expense" {
		finance.Type = "income"
	} else {
		finance.Type = "expense"
	}

	if err := b.db.Finances.UpdateFinance(finance.ID, finance); err != nil {
		return err
	}

	f_ext, err := b.detailedFinance(finance)
	if err != nil {
		return err
	}

	return b.constructListActions(c, f_ext, true)
}

func (b Bot) onEditAmout(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	if err := b.db.Users.SetState(c.Sender().ID, database.StateEditAmount); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "edit_amount"),
		b.Markup(c, "back_to_actions", finance),
	)
}

func (b Bot) onEditedAmount(c tele.Context) error {
	user := middle.User(c)

	finance, err := b.db.Finances.ByID(user.GetCache().FinanceID)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(c.Text(), 64)
	if err != nil {
		return err
	}
	finance.Amount = amount

	if err := b.db.Finances.UpdateFinance(finance.ID, finance); err != nil {
		return err
	}

	f_ext, err := b.detailedFinance(finance)
	if err != nil {
		return err
	}

	return b.constructListActions(c, f_ext, true)
}

func (b Bot) onEditCategory(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	if err := b.db.Users.SetState(c.Sender().ID, database.StateEditCategory); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "edit_category"),
		b.Markup(c, "back_to_actions", finance),
	)
}

func (b Bot) onEditedCategory(c tele.Context) error {
	var (
		user = middle.User(c)
		text = c.Text()
	)

	finance, err := b.db.Finances.ByID(user.GetCache().FinanceID)
	if err != nil {
		return err
	}

	if ok := regexp.MustCompile(`\d`).MatchString(text); ok {
		return c.Send(
			b.Text(c, "error_number"),
			tele.ForceReply,
		)
	}
	finance.Category = text

	if err := b.db.Finances.UpdateFinance(finance.ID, finance); err != nil {
		return err
	}

	f_ext, err := b.detailedFinance(finance)
	if err != nil {
		return err
	}

	return b.constructListActions(c, f_ext, true)
}

func (b Bot) onEditSubcategory(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	if err := b.db.Users.SetState(c.Sender().ID, database.StateEditSubCategory); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "edit_subcategory", finance),
		b.Markup(c, "back_to_actions", finance),
	)
}

func (b Bot) onEditedSubcategory(c tele.Context) error {
	text := c.Text()

	finance, err := b.db.Finances.ByID(middle.User(c).GetCache().FinanceID)
	if err != nil {
		return err
	}

	if ok := regexp.MustCompile(`\d`).MatchString(text); ok {
		return c.Send(
			b.Text(c, "error_number"),
			tele.ForceReply,
		)
	}
	finance.Subcategory = text

	if err := b.db.Finances.UpdateFinance(finance.ID, finance); err != nil {
		return err
	}

	f_ext, err := b.detailedFinance(finance)
	if err != nil {
		return err
	}

	return b.constructListActions(c, f_ext, true)
}

func (b Bot) onEditRecipient(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	if err := b.db.Users.SetState(c.Sender().ID, database.StateEditRecipient); err != nil {
		return err
	}

	return c.EditOrSend(
		b.Text(c, "edit_recipient"),
		b.Markup(c, "back_to_actions", finance),
	)
}

func (b Bot) onEditedRecipient(c tele.Context) error {
	var (
		text      = c.Text()
		user      = middle.User(c)
		media     = c.Message().Media()
		mediaID   = media.MediaFile().FileID
		mediaType = media.MediaType()
	)

	c.Delete()

	finance, err := b.db.Finances.ByID(middle.User(c).GetCache().FinanceID)
	if err != nil {
		return err
	}
	finance.Type = strings.Title(finance.Type)

	_, err = b.db.Recipients.ByID(finance.ID)
	r := database.Recipient{
		FinanceID:    finance.ID,
		Media:        mediaID,
		MediaType:    mediaType,
		MediaCaption: text,
	}

	if err != nil {
		if err == sql.ErrNoRows {
			if err := b.db.Recipients.Add(r); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := b.db.Recipients.UpdateRecipient(r); err != nil {
		return err
	}

	f_ext := Finance{
		Finance:   finance,
		Recipient: r,
	}

	b.Delete(user.ListMessage())
	b.Delete(user.ListActionsMessage())

	return b.constructListActions(c, f_ext)
}

func (b Bot) onBackToFinanceActions(c tele.Context) error {
	id, _ := strconv.Atoi(c.Data())

	finance, err := b.db.Finances.ByID(id)
	if err != nil {
		return err
	}

	f_ext := Finance{Finance: finance}

	_, err = b.Edit(
		middle.User(c).ListActionsMessage(),
		b.Text(c, "list_actions", f_ext),
		b.Markup(c, "list_opts", f_ext),
	)
	return err
}

func (b Bot) ShowAlert(c tele.Context) error {
	return c.Respond(
		&tele.CallbackResponse{
			Text:      b.Text(c, "list_single_record"),
			ShowAlert: true,
		},
	)
}

func (b Bot) detailedFinance(finance database.Finance) (Finance, error) {
	finance.Type = strings.Title(finance.Type)

	r, err := b.db.Recipients.ByID(finance.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return Finance{Finance: finance}, nil
		}
		return Finance{}, err
	}

	f_ext := Finance{Finance: finance, Recipient: r}

	return f_ext, nil
}

func (b Bot) financeExt(c tele.Context, page int) (Finance, error) {
	user := middle.User(c)
	user.UpdateCache("ListPage", page)
	b.db.Users.SetCache(user)

	finance, err := b.db.Finances.UserByOffset(user)
	if search, ok := searchPref.Load(c.Sender().ID); ok {
		finance, err = b.db.Finances.SearchByOffset(user, search)
	}
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
		msgActions = user.ListActionsMessage()
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
		if finance.Media == "" {
			_, err = b.Edit(
				msgList,
				what,
			)
		} else {
			if _, err := b.EditCaption(msgList, b.Text(c, "list_ext", finance)); err != nil {
				if err == tele.ErrSameMessageContent && err == tele.NewError(400, "", "there is no caption in the message to edit") {
					_, err = b.Edit(
						msgList,
						what,
					)
				}
			}
		}
		if err != nil {
			return err
		}
	} else {
		msgList, err = b.Send(
			c.Sender(),
			what,
		)
	}

	if editing {
		_, err = b.Edit(
			msgActions,
			b.Text(c, "list_actions", finance),
			b.actionMarkup(c, finance),
		)
		return err
	}
	msgActions, err = b.Send(
		c.Sender(),
		b.Text(c, "list_actions", finance),
		b.actionMarkup(c, finance),
	)
	if err != nil {
		return err
	}

	user.UpdateCache("ListMessageID", msgList.ID)
	user.UpdateCache("FinanceID", finance.Finance.ID)
	user.UpdateCache("ActionsMessageID", msgActions.ID)
	return b.db.Users.SetCache(user)
}

func (b Bot) actionMarkup(c tele.Context, finance Finance) *tele.ReplyMarkup {
	markup := b.Markup(c, "list_opts", finance)

	_, ok := searchPref.Load(c.Sender().ID)
	if !ok {
		return markup
	}

	markup.InlineKeyboard = markup.InlineKeyboard[1:]

	return markup
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
	case "video":
		return &tele.Video{
			File:    tele.File{FileID: f.Media},
			Caption: b.Text(c, "list_ext", f),
		}
	case "animation":
		return &tele.Animation{
			File:    tele.File{FileID: f.Media},
			Caption: b.Text(c, "list_ext", f),
		}
	default:
		return nil
	}
}
