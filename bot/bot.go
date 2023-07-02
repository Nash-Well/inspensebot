package bot

import (
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"inspense-bot/bot/middle"
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

	if err = b.SetCommands(lt.Commands()); err != nil {
		return nil, err
	}

	return &Bot{
		Bot:    b,
		Layout: lt,
		db:     boot.DB,
	}, nil
}

func (b *Bot) Start() {
	// Middleware
	b.Use(middle.SetUser(b.db))
	b.Use(b.Middleware("uk", b.localeFunc))

	// Handlers
	b.Handle("/start", b.onStart)
	b.Handle("/add", b.onAdd)
	b.Handle(tele.OnText, b.onText)
	b.Handle(tele.OnMedia, b.onMedia)

	// Callbacks
	b.Handle(b.Callback("lang"), b.onLanguage)
	b.Handle(b.Callback("fin_type"), b.onType)
	b.Handle(b.Callback("category"), b.onQuickCategory)
	b.Handle(b.Callback("subcat"), b.onSubMenu)
	b.Handle(b.Callback("recipient"), b.onRecipient)
	b.Handle(b.Callback("cancel"), b.onCancel)
	b.Handle(b.Callback("back"), b.onBackCategory)
	b.Handle(b.Callback("forward"), b.onForwardCategory)

	for _, locale := range b.Locales() {
		b.Handle(b.ButtonLocale(locale, "add"), b.onAdd)
		b.Handle(b.ButtonLocale(locale, "settings"), b.onSettings)
	}

	b.Bot.Start()
}
