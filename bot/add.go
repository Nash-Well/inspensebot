package bot

import tele "gopkg.in/telebot.v3"

func (b Bot) onAdd(c tele.Context) error {
	return c.Send(
		b.Text(c, "add_fin"),
		b.Markup(c, "type_menu"),
	)
}
