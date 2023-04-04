package bot

import (
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"inspense-bot/database"
)

type Bot struct {
	*tele.Bot
	*layout.Layout
	db *database.DB
}

func New(path string, boot BootStrap) (*Bot, error) {
	lt, err := layout.New(path)
	if err != nil {
		return nil, err
	}

	b, err := tele.NewBot(lt.Settings())
	if err != nil {
		return nil, err
	}

	if cmds, err := b.Commands(); err != nil {
		if err = b.SetCommands(cmds); err != nil {
			return nil, err
		}
	}

	return &Bot{
		Bot:    b,
		Layout: lt,
		db:     boot.DB,
	}, nil
}

func (b *Bot) Start() {
	// Middleware
	b.Use(b.Middleware("uk", b.localeFunc))

	// Handlers
	b.Handle("/start", b.onStart)

	// Callbacks
	b.Handle(b.Callback("lang"), b.onLanguage)

	for _, locale := range b.Locales() {
		b.Handle(b.ButtonLocale(locale, "add"), b.onAdd)
		b.Handle(b.ButtonLocale(locale, "settings"), b.onSettings)
	}

	b.Bot.Start()
}
