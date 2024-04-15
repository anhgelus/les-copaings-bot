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
		if i != l {
			chans += fmt.Sprintf("> <#%s>", c)
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
		cfg.Save()
	case "del":
		_, r := cfg.FindXpRole(role.ID)
		if r == nil {
			err := resp.Message("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				utils.SendAlert("commands/config.go - Role not found (del)", err.Error())
			}
			return
		}
		gokord.DB.Delete(r)
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
		gokord.DB.Save(r)
	default:
		err := resp.Message("Le type d'action n'est pas valide.").Send()
		if err != nil {
			utils.SendAlert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
	err := resp.Message("La configuration a bien été mise à jour.").Send()
	if err != nil {
		utils.SendAlert("commands/config.go - Config updated", err.Error())
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
			utils.SendAlert("commands/config.go - Channel not set", err.Error())
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
	cfg.Save()
	err := resp.Message("Modification sauvegardé.").Send()
	if err != nil {
		utils.SendAlert("commands/config.go - Modification saved message", err.Error())
	}
}
