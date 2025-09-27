package commands

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"
)

func Credits(s bot.Session, _ *event.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
	msg := "**Les Copaings**, le bot gérant les serveurs privés de [anhgelus](<https://anhgelus.world/>).\n"
	msg += "Code source : <https://git.anhgelus.world/anhgelus/les-copaings-bot>\n\n"
	msg += "Host du bot : " + gokord.BaseCfg.GetAuthor() + ".\n\n"
	msg += "Utilise :\n- [anhgelus/gokord](<https://github.com/anhgelus/gokord>)\n"
	msg += "- [Inter](<https://github.com/rsms/inter>)"
	err := resp.SetMessage(msg).Send()
	if err != nil {
		s.Logger().Error("sending credits", "error", err)
	}
}
