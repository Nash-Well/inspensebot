package bot

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/puzpuzpuz/xsync/v2"
	tele "gopkg.in/telebot.v3"
	"inspense-bot/bot/middle"
	"inspense-bot/bot/types"
	"inspense-bot/pkg/excel"
	"strconv"
	"strings"
)

var searchPref = xsync.NewIntegerMapOf[int64, types.Search]()

func (b Bot) onListType(c tele.Context) error {
	var (
		userID    = c.Sender().ID
		msgMarkup = c.Message().ReplyMarkup.InlineKeyboard
	)

	search := types.Search{Type: c.Data(), Search: "year", IsReport: false}
	if len(msgMarkup) == 1 {
		search.IsReport = true
	}
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

	markup := b.dateMarkup(c, years, true)

	if search.IsReport {
		ml := len(markup.InlineKeyboard) - 1
		markup.InlineKeyboard[ml] = append(markup.InlineKeyboard[ml][:1], markup.InlineKeyboard[ml][2:]...)
	}

	_, err = b.Edit(
		middle.User(c).ListMessage(),
		b.Text(c, "search_year"),
		markup,
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

	if search.Search == "day" && search.IsReport {
		search.Search = "month"

		finances, err := b.db.Finances.Finances(userID, search)
		if err != nil {
			return err
		}

		data, err := excel.DetailedReport(finances, search.Type, true)
		if err != nil {
			return err
		}

		user := middle.User(c)
		b.Delete(user.ListMessage())

		_, err = b.Send(
			c.Sender(),
			&tele.Document{
				File:     tele.FromReader(bytes.NewReader(data)),
				FileName: b.Text(c, "report") + ".xlsx",
			},
		)

		searchPref.Delete(userID)
		return err
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
		tempSearch = search
		markup     = b.listMarkup(c)
		ml         = len(markup.InlineKeyboard) - 1
	)

	if search.IsReport {
		markup.InlineKeyboard = markup.InlineKeyboard[ml-1 : ml]
	}

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

	if search.IsReport {
		finances, err := b.db.Finances.Finances(userID, search)
		if err != nil {
			return err
		}

		data, err := excel.DetailedReport(finances, search.Type)
		if err != nil {
			return err
		}

		user := middle.User(c)
		b.Delete(user.ListMessage())

		_, err = b.Send(
			c.Sender(),
			&tele.Document{
				File:     tele.FromReader(bytes.NewReader(data)),
				FileName: b.Text(c, "report") + ".xlsx",
			},
		)

		searchPref.Delete(userID)
		return err
	}

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
