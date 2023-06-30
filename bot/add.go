package bot

import (
	"errors"
	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
	"math"
	"regexp"
	"strconv"
	"time"
)

var finCache = xsync.NewIntegerMapOf[int64, database.Finance]()

func (b Bot) onAdd(c tele.Context) error {
	if err := b.db.Users.SetState(c.Sender().ID, database.StateAddType); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "add_fin"),
		b.Markup(c, "type_menu"),
	)
}

func (b Bot) onType(c tele.Context) error {
	id := c.Sender().ID

	finCache.Store(
		id,
		database.Finance{Type: c.Data()},
	)
	defer c.Delete()

	if err := b.db.Users.SetState(id, database.StateAddAmount); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "add_amount"),
		tele.ForceReply,
	)
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

	if amount < 0 {
		return c.Send(
			b.Text(c, "error_negative"),
			tele.ForceReply,
		)
	}

	finance.UserID = userID
	finance.Amount = amount

	finCache.Store(userID, finance)

	if err := b.db.Users.SetState(userID, database.StateAddDate); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "add_date"),
		tele.ForceReply,
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
		return c.Send(
			b.Text(c, "error_date"),
			tele.ForceReply,
		)
	}

	if t.After(time.Now()) {
		return c.Send(
			b.Text(c, "error_time"),
			tele.ForceReply,
		)
	}

	finance.Date = t
	finCache.Store(userID, finance)

	if err := b.db.Users.SetState(userID, database.StateAddCategory); err != nil {
		return err
	}

	list, err := b.db.Finances.CategoryList(userID, 0)
	if err != nil {
		return err
	}

	if len(list) > 0 {
		return b.quickCategories(c, list, 0)
	}

	return c.Send(
		b.Text(c, "add_category"),
		tele.ForceReply,
	)
}

func (b Bot) quickCategories(c tele.Context, list []database.Finance, page int) error {
	var (
		markup          = b.NewMarkup()
		navMarkup       = b.Markup(c, "nav_bar", page)
		categoryButtons [][]tele.InlineButton
	)

	for _, v := range list {
		button := tele.InlineButton{
			Unique: "category",
			Data:   v.Category,
			Text:   v.Category,
		}

		categoryButtons = append(categoryButtons, []tele.InlineButton{button})
	}

	markup.InlineKeyboard = append(markup.InlineKeyboard, categoryButtons...)
	markup.InlineKeyboard = append(markup.InlineKeyboard, navMarkup.InlineKeyboard...)

	return c.EditOrSend(
		b.Text(c, "add_category"),
		markup,
	)
}

func (b Bot) onBackCategory(c tele.Context) error {
	var (
		page, _ = strconv.Atoi(c.Data())
		userID  = c.Sender().ID
	)

	page -= 1
	if page < 0 {
		count, err := b.db.Finances.CategoryCount(userID)
		if err != nil {
			return err
		}

		var (
			res = float64(count) / 4.0
			dec = math.Mod(res, 1.0)
		)
		if dec == 0.25 || dec == 0.5 || dec == 0.75 {
			res = math.Ceil(res)
		}
		page = int(res)
	}

	list, err := b.db.Finances.CategoryList(userID, page)
	if err != nil {
		return err
	}

	return b.quickCategories(c, list, page)
}

func (b Bot) onForwardCategory(c tele.Context) error {
	var (
		page, _ = strconv.Atoi(c.Data())
		userID  = c.Sender().ID
	)
	page += 1

	list, err := b.db.Finances.CategoryList(userID, page)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		list, err = b.db.Finances.CategoryList(userID, 0)
		if err != nil {
			return err
		}
		page = 0
	}

	return b.quickCategories(c, list, page)
}

func (b Bot) onQuickCategory(c tele.Context) error {
	var (
		userID   = c.Sender().ID
		category = c.Args()[0]
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	finance.Category = category
	finCache.Store(userID, finance)

	defer c.Delete()

	if err := b.db.Users.SetState(userID, database.StateAddSubCategory); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "subcategory"),
		b.Markup(c, "subcategory_menu"),
	)

}

func (b Bot) onCategory(c tele.Context) error {
	var (
		userID = c.Sender().ID
		msg    = c.Text()
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	if ok := regexp.MustCompile(`\d`).MatchString(msg); ok {
		return c.Send(
			b.Text(c, "error_number"),
			tele.ForceReply,
		)
	}

	finance.Category = msg
	finCache.Store(userID, finance)

	if err := b.db.Users.SetState(userID, database.StateAddSubCategory); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "subcategory"),
		b.Markup(c, "subcategory_menu"),
		tele.ForceReply,
	)
}

func (b Bot) onSubCategory(c tele.Context) error {
	var (
		userID = c.Sender().ID
		msg    = c.Text()
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	if ok := regexp.MustCompile(`\d`).MatchString(msg); ok {
		return c.Send(
			b.Text(c, "error_number"),
			tele.ForceReply,
		)
	}

	finance.Subcategory = msg
	finCache.Store(userID, finance)

	return b.addFinance(c, finance)
	//return c.EditOrSend(
	//	b.Text(c, "recipient_exists"),
	//	b.Markup(c, "recipient_menu"),
	//)
}

func (b Bot) onSubMenu(c tele.Context) error {
	data := c.Data()
	finance, err := userCache(c.Sender().ID)
	if err != nil {
		return err
	}

	defer c.Delete()

	switch data {
	case "approval":
		return c.Send(
			b.Text(c, "add_subcategory"),
			tele.ForceReply,
		)
	case "not_apr":
		return b.addFinance(c, finance)
	default:
		return nil
	}
}

func (b Bot) addFinance(c tele.Context, f database.Finance) error {
	var (
		userID = c.Sender().ID
	)

	if err := b.db.Finances.Create(f); err != nil {
		return err
	}

	if err := b.db.Users.SetState(userID, database.StateIdle); err != nil {
		return err
	}

	return c.Send(b.Text(c, "fin_added"))
}

func userCache(userID int64) (database.Finance, error) {
	finance, ok := finCache.Load(userID)
	if !ok {
		return database.Finance{}, errors.New("no such finance in cache")
	}

	return finance, nil
}
