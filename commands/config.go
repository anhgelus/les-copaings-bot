package commands

import (
	"fmt"
	"slices"
	"strings"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

const (
	ConfigModify = "config_modify"
	OpenConfig   = "config"
)

func ConfigResponse(i *event.InteractionCreate) *interaction.Response {
	cfg := config.GetGuildConfig(i.GuildID)
	roles := ""
	l := len(cfg.XpRoles) - 1
	slices.SortFunc(cfg.XpRoles, func(xp1, xp2 config.XpRole) int {
		return int(xp2.XP) - int(xp1.XP)
	})
	for i, r := range cfg.XpRoles {
		if i == l {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>", exp.Level(r.XP), r.RoleID)
		} else {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>\n", exp.Level(r.XP), r.RoleID)
		}
	}
	if len(roles) == 0 {
		roles = "Aucun rôle configuré"
	}
	disChans := strings.Split(cfg.DisabledChannels, ";")
	var disChansDefault []component.SelectMenuDefaultValue
	for _, c := range disChans {
		if c != "" {
			disChansDefault = append(disChansDefault, component.SelectMenuDefaultValue{
				ID:   c,
				Type: types.SelectMenuDefaultValueChannel,
			})
		}
	}
	var defaultChan []component.SelectMenuDefaultValue
	if len(cfg.FallbackChannel) > 0 {
		defaultChan = append(defaultChan, component.SelectMenuDefaultValue{
			ID:   cfg.FallbackChannel,
			Type: types.SelectMenuDefaultValueChannel,
		})
	}
	zero := 0
	content := []component.Component{
		&component.Container{
			Components: []component.Message{
				&component.TextDisplay{Content: "## Configuration"},
				&component.Separator{},
				&component.TextDisplay{Content: "**Salons par défaut**\n-# Les niveaux obtenue grâce à un appel sont affichés ici"},
				&component.ActionsRow{
					Components: []component.Message{
						&component.SelectMenu{
							MenuType:      types.SelectMenuChannel,
							CustomID:      config.ModifyFallbackChannel,
							Placeholder:   "Pas de salon par défaut",
							MinValues:     &zero,
							MaxValues:     1,
							DefaultValues: defaultChan,
						},
					},
				},
				&component.TextDisplay{Content: "**Salons désactivé**\n-# Les messages ne donneront pas d'expérience dans ces salons"},
				&component.ActionsRow{
					Components: []component.Message{
						&component.SelectMenu{
							MenuType:      types.SelectMenuChannel,
							CustomID:      config.ModifyDisChannel,
							Placeholder:   "Pas de salons désactivé",
							MinValues:     &zero,
							MaxValues:     25,
							DefaultValues: disChansDefault,
						},
					},
				},
				&component.Section{
					Components: []component.Message{
						&component.TextDisplay{Content: "**Rôles de niveau**\n" + roles},
					},
					Accessory: &component.Button{
						Label:    "Modifier",
						Style:    component.ButtonStyleSecondary,
						CustomID: config.ModifyXpRole,
					},
				},
				&component.Section{
					Components: []component.Message{
						&component.TextDisplay{
							Content: fmt.Sprintf("**Jours avant la réduction**\n-# Seule l'expérience gagnée les x derniers jours est comptabilisée dans le niveau par défaut\n%d jours", cfg.DaysXPRemains),
						},
					},
					Accessory: &component.Button{
						Label:    "Modifier",
						Style:    component.ButtonStyleSecondary,
						CustomID: config.ModifyTimeReduce,
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
	s bot.Session,
	i *event.InteractionCreate,
	_ cmd.OptionMap,
	resp *cmd.ResponseBuilder,
) {
	err := s.InteractionAPI().Respond(i.Interaction, ConfigResponse(i))

	if err != nil {
		s.Logger().Error("sending config", "error", err)
	}
}

func ConfigMessageComponent(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	_ *cmd.ResponseBuilder,
) {
	response := ConfigResponse(i)
	response.Type = types.InteractionResponseUpdateMessage
	err := s.InteractionAPI().Respond(i.Interaction, response)

	if err != nil {
		s.Logger().Error("sending config", "error", err)
	}
}

func ConfigModal(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.ModalSubmitData,
	_ *cmd.ResponseBuilder,
) {
	response := ConfigResponse(i)
	response.Type = types.InteractionResponseUpdateMessage
	err := s.InteractionAPI().Respond(i.Interaction, response)

	if err != nil {
		s.Logger().Error("sending config", "error", err)
	}
}
