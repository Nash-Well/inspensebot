package bot

import (
	"errors"
	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
	"regexp"
	"strconv"
	"time"
)

var (
	finCache   = xsync.NewIntegerMapOf[int64, database.Finance]()
	categories = xsync.NewMapOf[database.Finance]()
)

func (b Bot) onAdd(c tele.Context) error {
	if err := b.db.Users.SetState(c.Sender().ID, database.AddingTypeState); err != nil {
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

	if err := b.db.Users.SetState(id, database.AddingAmount); err != nil {
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

	if err := b.db.Users.SetState(userID, database.AddingDate); err != nil {
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

	if err := b.db.Users.SetState(userID, database.AddingCategory); err != nil {
		return err
	}

	list, err := b.db.Finances.CategoryList(userID)
	if err != nil {
		return err
	}

	if len(list) > 0 {
		return b.quickCategories(c, list)
	}

	return c.Send(
		b.Text(c, "add_category"),
		tele.ForceReply,
	)
}

func (b Bot) quickCategories(c tele.Context, list []database.Finance) error {
	var row tele.Row

	for _, val := range list {
		categories.Store(val.Category, val)
	}

	categories.Range(func(key string, f database.Finance) bool {
		if len(row) >= 6 {
			return false
		}

		row = append(row, *b.Button(c, "category", f))
		return true
	})

	var (
		markup = b.NewMarkup()
		navBtn = []tele.Btn{
			*b.Button(c, "back"),
			*b.Button(c, "forward"),
		}
	)

	row = append(row, navBtn...)
	markup.Inline(markup.Split(2, row)...)

	return c.Send(
		b.Text(c, "add_category"),
		markup,
	)
}

func (b Bot) onQuickCategory(c tele.Context) error {
	var (
		userID   = c.Sender().ID
		category = c.Args()[1]
	)

	finance, err := userCache(userID)
	if err != nil {
		return err
	}

	finance.Category = category
	finCache.Store(userID, finance)

	defer c.Delete()

	if err := b.db.Users.SetState(userID, database.AddingSubcategory); err != nil {
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

	if err := b.db.Users.SetState(userID, database.AddingSubcategory); err != nil {
		return err
	}

	return c.Send(
		b.Text(c, "subcategory"),
		//tele.ForceReply,
		b.Markup(c, "subcategory_menu"),
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

	if err := b.db.Users.SetState(userID, database.DefaultState); err != nil {
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
