package commands

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func ConfigShow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg := config.GetGuildConfig(i.GuildID)
	resp := utils.ResponseBuilder{C: s, I: i}
	roles := ""
	l := len(cfg.XpRoles) - 1
	for i, r := range cfg.XpRoles {
		if i == l {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>", xp.Level(r.XP), r.RoleID)
		} else {
			roles += fmt.Sprintf("> Niveau %d - <@&%s>\n", xp.Level(r.XP), r.RoleID)
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
	err := resp.Embeds([]*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Config",
			Description: "Configuration",
			Color:       utils.Success,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Salons par défaut",
					Value:  fmt.Sprintf("<#%s>", cfg.FallbackChannel),
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
			},
		},
	}).Send()
	if err != nil {
		utils.SendAlert("config/guild.go - Sending config", err.Error())
	}
}

func ConfigXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optMap := utils.GenerateOptionMapForSubcommand(i)
	resp := utils.ResponseBuilder{C: s, I: i}
	resp.IsEphemeral()
	// verify every args
	t, ok := optMap["type"]
	if !ok {
		err := resp.Message("Le type d'action n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Action type not set", err.Error())
		}
		return
	}
	ts := t.StringValue()
	lvl, ok := optMap["level"]
	if !ok {
		err := resp.Message("Le niveau n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Level not set", err.Error())
		}
		return
	}
	level := lvl.IntValue()
	if level < 1 {
		err := resp.Message("Le niveau doit forcément être supérieur à 0.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid level", err.Error())
		}
		return
	}
	exp := xp.XPForLevel(uint(level))
	r, ok := optMap["role"]
	if !ok {
		err := resp.Message("Le rôle n'a pas été renseigné.").Send()
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
				err := resp.Message("Le rôle est déjà présent dans la config").Send()
				if err != nil {
					utils.SendAlert("commands/config.go - Role already in config", err.Error())
				}
				return
			}
		}
		cfg.XpRoles = append(cfg.XpRoles, config.XpRole{
			XP:     exp,
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
			err := resp.Message("Le rôle n'a pas été trouvé dans la config.").Send()
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
			err := resp.Message("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Role not found (edit)", err.Error())
			}
			return
		}
		r.XP = exp
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
		err := resp.Message("Le type d'action n'est pas valide.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
	if err != nil {
		err = resp.Message("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.Message("La configuration a bien été mise à jour.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Config updated message", err.Error())
	}
}

func ConfigChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optMap := utils.GenerateOptionMapForSubcommand(i)
	resp := utils.ResponseBuilder{C: s, I: i}
	resp.IsEphemeral()
	// verify every args
	t, ok := optMap["type"]
	if !ok {
		err := resp.Message("Le type d'action n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Action type not set", err.Error())
		}
		return
	}
	ts := t.StringValue()
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.Message("Le salon n'a pas été renseigné.").Send()
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
			err := resp.Message("Le salon est déjà dans la liste des salons désactivés").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Channel already disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels += channel.ID + ";"
	case "del":
		if !strings.Contains(cfg.DisabledChannels, channel.ID) {
			err := resp.Message("Le salon n'est pas désactivé").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Channel not disabled", err.Error())
			}
			return
		}
		cfg.DisabledChannels = strings.ReplaceAll(cfg.DisabledChannels, channel.ID+";", "")
	default:
		err := resp.Message("Le type d'action n'est pas valide.").Send()
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
		err = resp.Message("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.Message("Modification sauvegardé.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Modification saved message", err.Error())
	}
}

func ConfigFallbackChannel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optMap := utils.GenerateOptionMapForSubcommand(i)
	resp := utils.ResponseBuilder{C: s, I: i}
	resp.IsEphemeral()
	// verify every args
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.Message("Le salon n'a pas été renseigné.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Channel not set (fallback)", err.Error())
		}
		return
	}
	channel := salon.ChannelValue(s)
	if channel.Type != discordgo.ChannelTypeGuildText {
		err := resp.Message("Le salon n'est pas un salon textuel.").Send()
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
		err = resp.Message("Il y a eu une erreur lors de la modification de de la base de données.").Send()
	} else {
		err = resp.Message("Salon enregistré.").Send()
	}
	if err != nil {
		utils.SendAlert("commands/config.go - Channel saved message", err.Error())
	}
}
