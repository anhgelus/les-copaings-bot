package commands

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"strings"
)

const (
	SelectConfigModify             = "config_modify"
	SelectOptConfigXpRole          = "xp_role"
	SelectOptConfigDisChannel      = "disabled_channel"
	SelectOptConfigFallbackChannel = "fallback_channel"
	SelectOptConfigTimeReduce      = "time_reduce"
)

func Config(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
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
		Color: utils.Success,
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
	}).AddComponent(discordgo.ActionsRow{Components: []discordgo.MessageComponent{
		discordgo.SelectMenu{
			MenuType:    discordgo.StringSelectMenu,
			CustomID:    SelectConfigModify,
			Placeholder: "Modifier...",
			Options: []discordgo.SelectMenuOption{
				{
					Label:       "Rôles liés à l'XP",
					Value:       SelectOptConfigXpRole,
					Description: "Gère les rôles liés à l'XP",
					Emoji:       &discordgo.ComponentEmoji{Name: "🏅"},
				},
				{
					Label:       "Salons désactivés",
					Value:       SelectOptConfigDisChannel,
					Description: "Gère les salons désactivés",
					Emoji:       &discordgo.ComponentEmoji{Name: "❌"},
				},
				{
					Label:       "Salons de repli", // I don't have a better idea for this...
					Value:       SelectOptConfigFallbackChannel,
					Description: "Spécifie le salon de repli",
					Emoji:       &discordgo.ComponentEmoji{Name: "💾"},
				},
				{
					Label:       "Temps avec la réduction",
					Value:       SelectOptConfigTimeReduce,
					Description: "Gère le temps avant la réduction d'XP",
					Emoji:       &discordgo.ComponentEmoji{Name: "⌛"},
				},
			},
			Disabled: false,
		},
	}}).IsEphemeral().Send()
	if err != nil {
		utils.SendAlert("config/guild.go - Sending config", err.Error())
	}
}

func ConfigXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	resp.IsEphemeral()
	// verify every args
	t, ok := optMap["type"]
	if !ok {
		err := resp.SetMessage("Le type d'action n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Action type not set", err.Error())
		}
		return
	}
	ts := t.StringValue()
	lvl, ok := optMap["level"]
	if !ok {
		err := resp.SetMessage("Le niveau n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Level not set", err.Error())
		}
		return
	}
	level := lvl.IntValue()
	if level < 1 {
		err := resp.SetMessage("Le niveau doit forcément être supérieur à 0.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid level", err.Error())
		}
		return
	}
	xp := exp.LevelXP(uint(level))
	r, ok := optMap["role"]
	if !ok {
		err := resp.SetMessage("Le rôle n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Role not set", err.Error())
		}
		return
	}
	role := r.RoleValue(s, i.GuildID)
	cfg := config.GetGuildConfig(i.GuildID)

	// add or delete or edit
	var err error
	switch ts {
	case "add":
		for _, r := range cfg.XpRoles {
			if r.RoleID == role.ID {
				err = resp.SetMessage("Le rôle est déjà présent dans la config").Send()
				if err != nil {
					utils.SendAlert("commands/config.go - Role already in config", err.Error())
				}
				return
			}
		}
		cfg.XpRoles = append(cfg.XpRoles, config.XpRole{
			XP:     xp,
			RoleID: role.ID,
		})
		err = cfg.Save()
		if err != nil {
			utils.SendAlert(
				"commands/config.go - Saving config",
				err.Error(),
				"guild_id",
				i.GuildID,
				"role_id",
				role.ID,
				"type",
				"add",
			)
		}
	case "del":
		_, r := cfg.FindXpRole(role.ID)
		if r == nil {
			err = resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Role not found (del)", err.Error())
			}
			return
		}
		err = gokord.DB.Delete(r).Error
		if err != nil {
			utils.SendAlert(
				"commands/config.go - Deleting entry",
				err.Error(),
				"guild_id",
				i.GuildID,
				"role_id",
				role.ID,
				"type",
				"del",
			)
		}
	case "edit":
		_, r := cfg.FindXpRole(role.ID)
		if r == nil {
			err = resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Role not found (edit)", err.Error())
			}
			return
		}
		r.XP = xp
		err = gokord.DB.Save(r).Error
		if err != nil {
			utils.SendAlert(
				"commands/config.go - Saving config",
				err.Error(),
				"guild_id",
				i.GuildID,
				"role_id",
				role.ID,
				"type",
				"edit",
			)
		}
	default:
		err = resp.SetMessage("Le type d'action n'est pas valide.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
	if err != nil {
		err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.SetMessage("La configuration a bien été mise à jour.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Config updated message", err.Error())
	}
}

func ConfigChannel(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	t, ok := optMap["type"]
	if !ok {
		err := resp.SetMessage("Le type d'action n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Action type not set", err.Error())
		}
		return
	}
	ts := t.StringValue()
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.SetMessage("Le salon n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Channel not set (disabled)", err.Error())
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
				utils.SendAlert("commands/config.go - Channel already disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels += channel.ID + ";"
	case "del":
		if !strings.Contains(cfg.DisabledChannels, channel.ID) {
			err := resp.SetMessage("Le salon n'est pas désactivé").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Channel not disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels = strings.ReplaceAll(cfg.DisabledChannels, channel.ID+";", "")
	default:
		err := resp.SetMessage("Le type d'action n'est pas valide.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
	// save
	err := cfg.Save()
	if err != nil {
		utils.SendAlert(
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
		utils.SendAlert("commands/config.go - Modification saved message", err.Error())
	}
}

func ConfigFallbackChannel(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.SetMessage("Le salon n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Channel not set (fallback)", err.Error())
		}
		return
	}
	channel := salon.ChannelValue(s)
	if channel.Type != discordgo.ChannelTypeGuildText {
		err := resp.SetMessage("Le salon n'est pas un salon textuel.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid channel type", err.Error())
		}
		return
	}
	cfg := config.GetGuildConfig(i.GuildID)
	cfg.FallbackChannel = channel.ID
	// save
	err := cfg.Save()
	if err != nil {
		utils.SendAlert(
			"commands/config.go - Saving config",
			err.Error(),
			"guild_id",
			i.GuildID,
			"channel_id",
			channel.ID,
		)
		err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.SetMessage("Salon enregistré.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Channel saved message", err.Error())
	}
}

func ConfigPeriodBeforeReduce(s *discordgo.Session, i *discordgo.InteractionCreate, optMap utils.OptionMap, resp *utils.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	days, ok := optMap["days"]
	if !ok {
		err := resp.SetMessage("Le nombre de jours n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Days not set (fallback)", err.Error())
		}
		return
	}
	d := days.IntValue()
	if d < 30 {
		err := resp.SetMessage("Le nombre de jours est inférieur à 30.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Days < 30 (fallback)", err.Error())
		}
		return
	}
	// save
	cfg := config.GetGuildConfig(i.GuildID)
	cfg.DaysXPRemains = uint(d)
	err := cfg.Save()
	if err != nil {
		utils.SendAlert(
			"commands/config.go - Saving config",
			err.Error(),
			"guild_id",
			i.GuildID,
			"days",
			d,
		)
		err = resp.SetMessage("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.SetMessage("Nombre de jours enregistré.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Days saved message", err.Error())
	}
}
