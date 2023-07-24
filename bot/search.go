package bot

import (
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/bot/types"
	"strconv"
	"strings"
)

var searchPref = xsync.NewIntegerMapOf[int64, types.Search]()

func (b Bot) onListType(c tele.Context) error {
	userID := c.Sender().ID

	search := types.Search{Type: c.Data(), Search: "year"}
	searchPref.Store(userID, search)

	years, err := b.db.Finances.FinanceDetailed(search, userID)
	if err != nil {
		return err
	}

	if len(years) == 0 {
		return c.Respond(&tele.CallbackResponse{
			Text: b.Text(c, "search_empty_type"),
		})
	}

	_, err = b.Edit(
		middle.User(c).ListMessage(),
		b.Text(c, "search_year"),
		b.dateMarkup(c, years, true),
	)

	return err
}

func (b Bot) onNumb(c tele.Context) error {
	var (
		userID      = c.Sender().ID
		userNumb, _ = strconv.Atoi(c.Data())
	)

	search, ok := searchPref.Load(userID)
	if !ok {
		return errors.New("bot/search: no such user")
	}

	switch search.Search {
	case "year":
		search.Year = userNumb
		search.Search = "month"
	case "month":
		search.Month = userNumb
		search.Search = "day"
	case "day":
		search.Day = userNumb
	}

	numbs, err := b.db.Finances.FinanceDetailed(search, userID)
	if err != nil {
		return err
	}

	searchPref.Store(userID, search)

	if search.Day != 0 {
		finance, err := b.financeExt(c, 0)
		if err != nil {
			return err
		}

		err = b.constructList(c, finance, true)
	} else {
		var markup *tele.ReplyMarkup

		if search.Search != "year" {
			markup = b.dateMarkup(c, numbs)
		} else {
			markup = b.dateMarkup(c, numbs, true)
		}

		_, err = b.Edit(
			middle.User(c).ListMessage(),
			b.Text(c, fmt.Sprintf("search_%s", search.Search)),
			markup,
		)
	}

	return err
}

func (b Bot) onSearchBack(c tele.Context) (err error) {
	var (
		userID = c.Sender().ID
		user   = middle.User(c)
	)

	search, ok := searchPref.Load(userID)
	if !ok {
		return errors.New("bot/search: no such user")
	}

	var (
		markup     = b.listMarkup(c)
		tempSearch = search
	)

	if search.Search != "year" {
		switch search.Search {
		case "month":
			tempSearch.Search = "year"
			tempSearch.Year = 0
		case "day":
			tempSearch.Search = "month"
			tempSearch.Month = 0
		}

		numbs, err := b.db.Finances.FinanceDetailed(tempSearch, userID)
		if err != nil {
			return err
		}

		if tempSearch.Search != "year" {
			markup = b.dateMarkup(c, numbs)
		} else {
			markup = b.dateMarkup(c, numbs, true)
		}
	}

	switch search.Search {
	case "year":
		_, err = b.Edit(
			user.ListMessage(),
			b.Text(c, "search_type"),
			markup,
		)
	case "month":
		_, err = b.Edit(
			user.ListMessage(),
			b.Text(c, "search_year"),
			markup,
		)
	case "day":
		_, err = b.Edit(
			user.ListMessage(),
			b.Text(c, "search_month"),
			markup,
		)
	}

	search = tempSearch
	searchPref.Store(userID, search)

	return
}

func (b Bot) onSearchDone(c tele.Context) error {
	userID := c.Sender().ID

	search, ok := searchPref.Load(userID)
	if !ok {
		return errors.New("bot/search: no such user")
	}

	switch search.Search {
	case "year":
		search.Search = ""
	case "month":
		search.Search = "year"
	case "day":
		search.Search = "month"
	}

	searchPref.Store(userID, search)

	finance, err := b.financeExt(c, 0)
	if err != nil {
		return err
	}

	return b.constructList(c, finance, true)
}

func (b Bot) onViewAll(c tele.Context) error {
	searchPref.Delete(c.Sender().ID)

	finance, err := b.financeExt(c, 0)
	if err != nil {
		return err
	}

	return b.constructList(c, finance, true)
}

func (b Bot) dateMarkup(c tele.Context, numbs []int, year ...bool) *tele.ReplyMarkup {
	var (
		row    tele.Row
		rows   []tele.Row
		isYear = len(year) > 0 && year[0]
		markup = b.NewMarkup()
	)

	for _, numb := range numbs {
		row = append(row, *b.Button(c, "search_numb", numb))
	}

	if isYear {
		rows = markup.Split(1, row)
	} else {
		rows = markup.Split(3, row)
	}

	rows = append(rows, tele.Row{*b.Button(c, "search_back"), *b.Button(c, "search_done")})
	markup.Inline(rows...)

	return markup
}

func (b Bot) listMarkup(c tele.Context) (markup *tele.ReplyMarkup) {
	markup = b.Markup(c, "type_menu")

	for i, row := range markup.InlineKeyboard {
		for j, cell := range row {
			markup.InlineKeyboard[i][j].Unique = strings.Replace(cell.Unique, "type", "search_type", 1)
		}
	}

	markup.InlineKeyboard = append(markup.InlineKeyboard, []tele.InlineButton{*b.Button(c, "search_all").Inline()})

	return
}
