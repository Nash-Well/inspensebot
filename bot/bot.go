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
	b.Handle("/list", b.onList)
	b.Handle("/share", b.onShare)
	b.Handle("/view", b.onView)
	b.Handle("/deny", b.onDeny)

	b.Handle(tele.OnText, b.onText)
	b.Handle(tele.OnMedia, b.onMedia)

	// Callbacks
	b.Handle(b.Callback("lang"), b.onLanguage)

	b.Handle(b.Callback("fin_type"), b.onType)
	b.Handle(b.Callback("subcat"), b.onSubMenu)

	b.Handle(b.Callback("recipient"), b.onRecipient)
	b.Handle(b.Callback("cancel"), b.onCancel)

	b.Handle(b.Callback("category"), b.onQuickCategory)
	b.Handle(b.Callback("back"), b.onBackCategory)
	b.Handle(b.Callback("forward"), b.onForwardCategory)

	b.Handle(b.Callback("list_back"), b.onBackList)
	b.Handle(b.Callback("list_forward"), b.onForwardList)
	b.Handle(b.Callback("list_func"), b.onFunctions)
	b.Handle(b.Callback("back_to_list"), b.onBackToList)

	// View
	b.Handle(b.Callback("forward_user"), b.onUser)
	b.Handle(b.Callback("view_back"), b.onBackView)
	b.Handle(b.Callback("view_forward"), b.onForwardView)
	b.Handle(b.Callback("from_user"), b.onUserDeny)

	for _, locale := range b.Locales() {
		b.Handle(b.ButtonLocale(locale, "add"), b.onAdd)
		b.Handle(b.ButtonLocale(locale, "list"), b.onList)
		b.Handle(b.ButtonLocale(locale, "settings"), b.onSettings)
	}

	b.Bot.Start()
}
