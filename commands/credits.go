package commands

import (
	"context"

	"github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/interaction"
)

func Credits(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	msg := "**Les Copaings**, le bot gérant les serveurs privés de [anhgelus](<https://anhgelus.world/>).\n"
	msg += "Code source : <https://git.anhgelus.world/anhgelus/les-copaings-bot>\n\n"
	msg += "Host du bot : " + gokord.BaseCfg.GetAuthor() + ".\n\n"
	msg += "Utilise :\n- [anhgelus/gokord](<https://github.com/anhgelus/gokord>)\n"
	msg += "- [Inter](<https://github.com/rsms/inter>)"
	resp := interaction.NewMessageResponse().Message(msg).Response()
	err := dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("sending credits", "error", err)
	}
}
