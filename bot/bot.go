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
	b.Use(b.Middleware("uk"))

	//Handlers
	b.Handle("/start", b.onStart)

	b.Bot.Start()
}
