package bot

import (
	"database/sql"
	"golang.org/x/exp/slices"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/database"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var finType = []string{"income", "expense"}

func (b Bot) onSearch(c tele.Context) (err error) {
	var (
		args     = c.Args()
		la       = len(args)
		finances []database.Finance
	)

	defer c.Delete()

	var t string
	switch {
	case la == 0:
		fallthrough
	case la == 1 && slices.Contains(finType, args[0]):
		return c.Send(b.Text(c, "search_pref"))
	case la > 1:
		if !slices.Contains(finType, args[0]) {
			return c.Send(b.Text(c, "search_pref"))
		}

		t, err = parseTime(args[1])
		if err != nil {
			return c.Send(
				b.Text(c, "error_date"),
				tele.ForceReply,
			)
		}
	}

	switch sd := args[0]; {
	case la > 0 && la <= 1:
		t, err := parseTime(sd)
		if err != nil {
			return c.Send(
				b.Text(c, "error_date"),
				tele.ForceReply,
			)
		}

		finances, err = b.db.Finances.SearchByDate(t)
	case la > 1 && la <= 2:
		finances, err = b.db.Finances.SearchByTypeDate(sd, t)

	case la > 2:
		fc := args[2]

		if _, err := regexp.MatchString(`\d`, fc); err != nil {
			return c.Send(
				b.Text(c, "error_number"),
				tele.ForceReply,
			)
		}

		finances, err = b.db.Finances.SearchByAll(sd, t, fc)
	}

	if err != nil {
		return err
	}

	if len(finances) == 0 {
		return c.Send(b.Text(c, "search_none"))
	}

	var row tele.Row
	for idx, f := range finances {
		row = append(row, *b.Button(c, "search_res", Finance{Finance: f, Page: idx + 1}))
	}

	markup := b.NewMarkup()
	markup.Inline(markup.Split(3, row)...)

	return c.Send(b.Text(c, "search_res", finances), markup)
}

func (b Bot) onSearchRes(c tele.Context) error {
	fnID, _ := strconv.Atoi(c.Data())

	c.Delete()

	finance, err := b.db.Finances.ByID(fnID)
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

	return b.constructListActions(c, f_ext)

}

func parseTime(fd string) (string, error) {
	t, err := time.Parse("02.01.2006", fd)
	if err != nil {
		t, err = time.Parse("02.01", fd)
		if err != nil {

		}

		t = t.AddDate(2023, 0, 0)
	}

	return strings.TrimSuffix(t.String(), " UTC"), nil
}
