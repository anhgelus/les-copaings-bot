package commands

import (
	"fmt"
	"strings"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord/cmd"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/emoji"
	"github.com/nyttikord/gokord/interaction"
)

const (
	ConfigModify = "config_modify"
	OpenConfig   = "config"
)

func ConfigResponse(i *discordgo.InteractionCreate) *interaction.Response {
	cfg := config.GetGuildConfig(i.GuildID)
	roles := ""
	l := len(cfg.XpRoles) - 1
	for i, r := range cfg.XpRoles {
		if i == l {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>", exp.Level(r.XP), r.RoleID)
		} else {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>\n", exp.Level(r.XP), r.RoleID)
		}
	}
	if len(roles) == 0 {
		roles = "Aucun rôle configuré :("
	}
	disChans := strings.Split(cfg.DisabledChannels, ";")
	l = len(disChans) - 1
	chans := ""
	for i, c := range disChans {
		if i == l-1 {
			chans += fmt.Sprintf("> <#%s>", c)
		} else if i != l {
			chans += fmt.Sprintf("> <#%s>\n", c)
		}
	}
	if len(chans) == 0 {
		chans = "Aucun salon désactivé :)"
	}
	var defaultChan string
	if len(cfg.FallbackChannel) == 0 {
		defaultChan = "Pas de valeur"
	} else {
		defaultChan = fmt.Sprintf("<#%s>", cfg.FallbackChannel)
	}
	content := []component.Component{
		&component.Container{
			Components: []component.Message{
				&component.TextDisplay{Content: "## Configuration"},
				&component.Separator{},
				&component.TextDisplay{Content: "**Salon par défaut**\n" + defaultChan},
				&component.TextDisplay{Content: "**Rôles de niveau**\n" + roles},
				&component.TextDisplay{Content: "**Salons ignorés**\n" + chans},
				&component.TextDisplay{
					Content: fmt.Sprintf("**Jours avant la réduction**\n%d jours", cfg.DaysXPRemains),
				},
				&component.ActionsRow{
					Components: []component.Message{
						&component.SelectMenu{
							MenuType:    types.SelectMenuString,
							Placeholder: "Gestion des paramètres",
							CustomID:    ConfigModify,
							Options: []component.SelectMenuOption{
								{
									Label: "Salons par défaut",
									Value: config.ModifyFallbackChannel,
									Emoji: &emoji.Component{Name: "📣"},
								},
								{
									Label: "Rôles de niveaux",
									Value: config.ModifyXpRole,
									Emoji: &emoji.Component{Name: "🏅"},
								},
								{
									Label: "Salons ignorés",
									Value: config.ModifyDisChannel,
									Emoji: &emoji.Component{Name: "🫣"},
								},
								{
									Label: "Temps avant la réduction d'expérience",
									Value: config.ModifyTimeReduce,
									Emoji: &emoji.Component{Name: "📉"},
								},
							},
						},
					},
				},
			},
		},
	}
	return &interaction.Response{
		Type: types.InteractionResponseChannelMessageWithSource,
		Data: &interaction.ResponseData{
			Components: content,
			Flags:      channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
		},
	}
}

func ConfigCommand(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ cmd.OptionMap,
	resp *cmd.ResponseBuilder,
) {
	err := session.InteractionAPI().Respond(i.Interaction, ConfigResponse(i))

	if err != nil {
		session.LogError(err, "config/guild.go - Sending config")
	}
}

func ConfigMessageComponent(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ *interaction.MessageComponentData,
	_ *cmd.ResponseBuilder,
) {
	response := ConfigResponse(i)
	response.Type = types.InteractionResponseUpdateMessage
	err := session.InteractionAPI().Respond(i.Interaction, response)

	if err != nil {
		session.LogError(err, "sending config")
	}
}
