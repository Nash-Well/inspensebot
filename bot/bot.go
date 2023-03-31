package bot

import (
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type Bot struct {
	*tele.Bot
	*layout.Layout
}

func New(path string) (*Bot, error) {
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
	}, nil
}

func (b *Bot) Statrt() {
	b.Use(b.Middleware("uk"))

	b.Bot.Start()
}
