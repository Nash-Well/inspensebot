package bot

import (
	tele "gopkg.in/telebot.v3"
)

func (b Bot) onSettings(c tele.Context) error {
	return c.Send(
		b.Text(c, "help_lang"),
		b.Markup(c, "lang_menu"),
	)
}

func (b Bot) localeFunc(c tele.Recipient) string {
	locale, _ := b.db.Users.Language(c)
	return locale
}

func (b Bot) onLanguage(c tele.Context) error {
	lang := c.Data()

	if l, _ := b.Locale(c); l == lang {
		return c.Respond()
	}

	defer c.Delete()
	if err := b.db.Users.SetLanguage(c.Sender().ID, lang); err != nil {
		return err
	}

	b.Layout.SetLocale(c, lang)

	return c.Send(
		b.Text(c, "lang_success"),
		b.Markup(c, "menu"),
	)
}
