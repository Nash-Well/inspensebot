package bot

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"time"

	"inspense-bot/bot/middle"
	"inspense-bot/database"

	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
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
	userID := c.Sender().ID

	finCache.Store(
		userID,
		database.Finance{Type: c.Data()},
	)
	defer c.Delete()

	if err := b.db.Users.SetState(userID, database.StateAddAmount); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "add_amount"),
		tele.ForceReply,
	)
}

func (b Bot) onAmount(c tele.Context) error {
	userID := c.Sender().ID

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(c.Text(), 64)
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
		text   = c.Text()
		userID = c.Sender().ID
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	t, err := time.Parse("02.01.2006", text)
	if err != nil {
		t, err = time.Parse("02.01", text)
		if err != nil {
			return c.Send(
				b.Text(c, "error_date"),
				tele.ForceReply,
			)
		}

		t = t.AddDate(2023, 0, 0)
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

	list, err := b.categoryList(c, 0)
	if err != nil {
		return err
	}

	if len(list) > 0 {
		return b.construcCategory(c, list)
	}

	return c.Send(
		b.Text(c, "add_category"),
		tele.ForceReply,
	)
}

func (b Bot) quickCategoryMarkup(c tele.Context, categories []string) *tele.ReplyMarkup {
	var (
		markup          = b.NewMarkup()
		categoryButtons [][]tele.InlineButton
		page            = middle.User(c).GetCache().CategoryPage
		navMarkup       = b.Markup(c, "nav_bar", page)
	)

	for _, category := range categories {
		button := tele.InlineButton{
			Unique: "category",
			Data:   category,
			Text:   category,
		}

		categoryButtons = append(categoryButtons, []tele.InlineButton{button})
	}
	markup.InlineKeyboard = append(markup.InlineKeyboard, categoryButtons...)

	count, err := b.categoryCount(c.Sender().ID)
	if err != nil {
		return &tele.ReplyMarkup{}
	}

	if count > 4 {
		markup.InlineKeyboard = append(markup.InlineKeyboard, navMarkup.InlineKeyboard...)
	}

	return markup
}

func (b Bot) onBackCategory(c tele.Context) error {
	page, _ := strconv.Atoi(c.Data())

	page -= 1
	if page < 0 {
		count, err := b.categoryCount(c.Sender().ID)
		if err != nil {
			return err
		}

		var (
			res = float64(count) / 4.0
			dec = math.Mod(res, 1.0)
		)
		if dec == 0.25 || dec == 0.5 || dec == 0.75 {
			res = math.Floor(res)
		} else {
			res -= 1
		}

		page = int(res)
	}

	list, err := b.categoryList(c, page)
	if err != nil {
		return err
	}

	return b.construcCategory(c, list, true)
}

func (b Bot) onForwardCategory(c tele.Context) error {
	page, _ := strconv.Atoi(c.Data())
	page += 1

	list, err := b.categoryList(c, page)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		list, err = b.categoryList(c, 0)
	}

	return b.construcCategory(c, list, true)
}

func (b Bot) onQuickCategory(c tele.Context) error {
	userID := c.Sender().ID

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	finance.Category = c.Args()[0]
	finCache.Store(userID, finance)

	user := middle.User(c)
	user.DeleteFromCache("CategoryPage")
	user.DeleteFromCache("CategoryMessageID")
	b.db.Users.SetCache(*user)

	defer c.Delete()

	if err := b.db.Users.SetState(userID, database.StateIdle); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "subcategory"),
		b.Markup(c, "subcategory_menu"),
	)

}

func (b Bot) onCategory(c tele.Context) error {
	var (
		text   = c.Text()
		userID = c.Sender().ID
	)

	finance, err := userCache(userID)
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
	finCache.Store(userID, finance)

	user := middle.User(c)
	if user.Exists("CategoryPage") {
		b.Delete(user.CategoryMessage())

		user.DeleteFromCache("CategoryPage")
		user.DeleteFromCache("CategoryMessageID")
		b.db.Users.SetCache(*user)
	}

	if err := b.db.Users.SetState(userID, database.StateIdle); err != nil {
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
		text   = c.Text()
		userID = c.Sender().ID
	)

	finance, err := userCache(userID)
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
	finCache.Store(userID, finance)

	return c.EditOrSend(
		b.Text(c, "recipient_exists"),
		b.Markup(c, "recipient_menu"),
	)
}

func (b Bot) onSubMenu(c tele.Context) error {
	switch c.Data() {
	case "approval":
		c.Delete()

		b.db.Users.SetState(c.Sender().ID, database.StateAddSubCategory)
		return c.Send(
			b.Text(c, "add_subcategory"),
			tele.ForceReply,
		)
	case "not_apr":
		return c.EditOrSend(
			b.Text(c, "recipient_exists"),
			b.Markup(c, "recipient_menu"),
		)
	default:
		return nil
	}
}

func (b Bot) addFinance(c tele.Context, f database.Finance) error {
	if _, err := b.db.Finances.Create(f); err != nil {
		return err
	}

	if err := b.db.Users.SetState(c.Sender().ID, database.StateIdle); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "fin_added"),
		b.Markup(c, "menu"),
	)
}

func userCache(userID int64) (database.Finance, error) {
	finance, ok := finCache.Load(userID)
	if !ok {
		return database.Finance{}, errors.New("no such finance in cache")
	}

	return finance, nil
}

func (b Bot) categoryList(c tele.Context, page int) ([]string, error) {
	user := middle.User(c)
	user.UpdateCache("CategoryPage", page)
	b.db.Users.SetCache(*user)

	finance, err := userCache(c.Sender().ID)
	if err != nil {
		return nil, err
	}

	list, err := b.db.Finances.CategoryList(*user, finance)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (b Bot) categoryCount(userID int64) (int, error) {
	finance, err := userCache(userID)
	if err != nil {
		return 0, err
	}

	count, err := b.db.Finances.CategoryCount(finance)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (b Bot) construcCategory(c tele.Context, list []string, edit ...bool) (err error) {
	var (
		editing     = len(edit) > 0 && edit[0]
		user        = middle.User(c)
		msgCategory *tele.Message
	)

	if editing {
		msgCategory = user.CategoryMessage()
	} else {
		msgCategory = c.Message()
	}

	if editing {
		msgCategory, err = b.Edit(
			msgCategory,
			b.Text(c, "add_category"),
			b.quickCategoryMarkup(c, list),
		)
	} else {
		msgCategory, err = b.Send(
			c.Sender(),
			b.Text(c, "add_category"),
			b.quickCategoryMarkup(c, list),
		)
	}
	if err != nil {
		return err
	}

	user.UpdateCache("CategoryMessageID", msgCategory.ID)
	return b.db.Users.SetCache(*user)
}
