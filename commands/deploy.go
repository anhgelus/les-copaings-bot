package commands

import (
	"context"

	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/interaction"
)

var commands []*interaction.Command

func newCmd(name, description string) *interaction.Command {
	return &interaction.Command{
		Type:             types.CommandChat,
		Name:             name,
		Description:      description,
		Contexts:         &[]types.InteractionContext{types.InteractionContextGuild},
		IntegrationTypes: &[]types.IntegrationInstall{types.IntegrationInstallGuild},
	}
}

func newOption(tp types.CommandOption, name, description string) *interaction.CommandOption {
	return &interaction.CommandOption{
		Type:        tp,
		Name:        name,
		Description: description,
	}
}

func init() {
	adm := int64(discord.PermissionManageGuild)

	rank := newCmd("rank", "Affiche le niveau d'un copaing")
	rank.Options = append(rank.Options, newOption(types.CommandOptionUser, "copaing", "Le niveau du Copaing que vous souhaitez obtenir"))

	cfg := newCmd("config", "Modifie la config")
	cfg.DefaultMemberPermissions = &adm

	top := newCmd("top", "Copaings les plus actifs")

	reset := newCmd("reset", "Reset l'xp")
	reset.DefaultMemberPermissions = &adm

	resetUser := newCmd("reset-user", "Reset l'xp d'un utilisation")
	resetUser.DefaultMemberPermissions = &adm
	resetUser.Options = append(resetUser.Options, newOption(types.CommandOptionUser, "copaing", "Copaing à reset"))

	credits := newCmd("credits", "Affiche les crédits du bot")

	stats := newCmd("stats", "Affiche des stats :D")
	stats.Options = append(stats.Options,
		newOption(types.CommandOptionInteger, "jours", "Nombre de jours à afficher dans le graphique"),
		newOption(types.CommandOptionUser, "copaing", "Copaing à inspecter"),
	)

	rolereact := newCmd("rolereact", "Envoie un message permettant de récupérer des rôles grâce à des réactions")
	rolereact.DefaultMemberPermissions = &adm
	rolereact.Options = append(rolereact.Options, newOption(types.CommandOptionChannel, "salon", "Salon de destination du message"))

	modify := newCmd("Modifier", "")
	modify.DefaultMemberPermissions = &adm
	modify.Type = types.CommandMessage

	commands = []*interaction.Command{rank, cfg, top, reset, resetUser, credits, stats, rolereact, modify}
}

func Deploy(ctx context.Context, dg bot.Session) error {
	guildID := ""
	if common.IsDebug(ctx) {
		guildID = dg.GuildAPI().State.Guilds()[0]
		gld, err := dg.GuildAPI().State.Guild(guildID)
		if err != nil {
			return err
		}
		bot.Logger(ctx).Debug("using guild as debug", "guild", gld.Name)
	}
	_, err := dg.InteractionAPI().CommandBulkOverwrite(dg.SessionState().Application().ID, guildID, commands).Do(ctx)
	return err
}
