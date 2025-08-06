package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/component"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"strings"
)

const (
	ConfigModify = "config_modify"
)

func Config(_ *discordgo.Session, i *discordgo.InteractionCreate, _ cmd.OptionMap, resp *cmd.ResponseBuilder) {
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
	err := resp.AddEmbed(&discordgo.MessageEmbed{
		Type:  discordgo.EmbedTypeRich,
		Title: "Config",
		Color: 0x10E6AD,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Salon par défaut",
				Value:  defaultChan,
				Inline: false,
			},
			{
				Name:   "Rôles liés aux niveaux",
				Value:  roles,
				Inline: false,
			},
			{
				Name:   "Salons désactivés",
				Value:  chans,
				Inline: false,
			},
			{
				Name:   "Jours avant la réduction",
				Value:  fmt.Sprintf("%d", cfg.DaysXPRemains),
				Inline: false,
			},
		},
	}).SetComponents(component.New().Add(component.NewActionRow().Add(
		component.NewStringSelect(ConfigModify).SetPlaceholder("Modifier...").
			AddOption(
				component.NewSelectOption("Rôles liés à l'XP", config.ModifyXpRole).
					SetDescription("Gère les rôles liés à l'XP").
					SetEmoji(&discordgo.ComponentEmoji{Name: "🏅"}),
			).
			AddOption(
				component.NewSelectOption("Salons désactivés", config.ModifyDisChannel).
					SetDescription("Gère les salons désactivés").
					SetEmoji(&discordgo.ComponentEmoji{Name: "❌"}),
			).
			AddOption(
				// I don't have a better idea for this...
				component.NewSelectOption("Salons de repli", config.ModifyFallbackChannel).
					SetDescription("Spécifie le salon de repli").
					SetEmoji(&discordgo.ComponentEmoji{Name: "💾"}),
			).
			AddOption(
				component.NewSelectOption("Temps avec la réduction", config.ModifyTimeReduce).
					SetDescription("Gère le temps avant la réduction d'XP").
					SetEmoji(&discordgo.ComponentEmoji{Name: "⌛"}),
			),
	))).IsEphemeral().Send()
	if err != nil {
		logger.Alert("config/guild.go - Sending config", err.Error())
	}
}

func ConfigChannel(s *discordgo.Session, i *discordgo.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	t, ok := optMap["type"]
	if !ok {
		err := resp.SetMessage("Le type d'action n'a pas été renseigné.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Action type not set", err.Error())
		}
		return
	}
	ts := t.StringValue()
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.SetMessage("Le salon n'a pas été renseigné.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Channel not set (disabled)", err.Error())
		}
		return
	}
	channel := salon.ChannelValue(s)
	cfg := config.GetGuildConfig(i.GuildID)
	switch ts {
	case "add":
		if strings.Contains(cfg.DisabledChannels, channel.ID) {
			err := resp.SetMessage("Le salon est déjà dans la liste des salons désactivés").Send()
			if err != nil {
				logger.Alert("commands/config.go - Channel already disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels += channel.ID + ";"
	case "del":
		if !strings.Contains(cfg.DisabledChannels, channel.ID) {
			err := resp.SetMessage("Le salon n'est pas désactivé").Send()
			if err != nil {
				logger.Alert("commands/config.go - Channel not disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels = strings.ReplaceAll(cfg.DisabledChannels, channel.ID+";", "")
	default:
		err := resp.SetMessage("Le type d'action n'est pas valide.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
	// save
	err := cfg.Save()
	if err != nil {
		logger.Alert(
			"commands/config.go - Saving config",
			err.Error(),
			"guild_id",
			i.GuildID,
			"type",
			ts,
			"channel_id",
			channel.ID,
		)
		err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.SetMessage("Modification sauvegardé.").Send()
	}
	if err != nil {
		logger.Alert("commands/config.go - Modification saved message", err.Error())
	}
}
