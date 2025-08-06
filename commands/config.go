package commands

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/component"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
	"time"
)

const (
	ConfigModify                = "config_modify"
	ConfigModifyXpRole          = "xp_role"
	ConfigModifyDisChannel      = "disabled_channel"
	ConfigModifyFallbackChannel = "fallback_channel"
	ConfigModifyTimeReduce      = "time_reduce"
	// xp role related
	XpRoleAdd       = "xp_role_add"
	XpRoleAddLevel  = "xp_role_add_level"
	XpRoleAddRole   = "xp_role_add_role"
	XpRoleDel       = "xp_role_del"
	XpRoleDelRole   = "xp_role_del_role"
	XpRoleEdit      = "xp_role_edit"
	XpRoleEditLevel = "xp_role_edit_level"
	XpRoleEditRole  = "xp_role_edit_role"
)

var (
	configModifyMap = map[string]uint{}
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
	}).SetComponents(component.New().Add(new(component.ActionRow).Add(
		new(component.StringSelect).SetPlaceholder("Modifier...").
			AddOption(
				component.NewSelectOption("Rôles liés à l'XP", ConfigModifyXpRole).
					SetDescription("Gère les rôles liés à l'XP").
					SetEmoji(&discordgo.ComponentEmoji{Name: "🏅"}),
			).
			AddOption(
				component.NewSelectOption("Salons désactivés", ConfigModifyDisChannel).
					SetDescription("Gère les salons désactivés").
					SetEmoji(&discordgo.ComponentEmoji{Name: "❌"}),
			).
			AddOption(
				component.NewSelectOption("Salons de repli", ConfigModifyFallbackChannel). // I don't have a better idea for this...
					SetDescription("Spécifie le salon de repli").
					SetEmoji(&discordgo.ComponentEmoji{Name: "💾"}),
			).
			AddOption(
				component.NewSelectOption("Temps avec la réduction", ConfigModifyTimeReduce).
					SetDescription("Gère le temps avant la réduction d'XP").
					SetEmoji(&discordgo.ComponentEmoji{Name: "⌛"}),
			).
			SetCustomID(ConfigModify),
	))).IsEphemeral().Send()
	if err != nil {
		logger.Alert("config/guild.go - Sending config", err.Error())
	}
}

func ConfigXP(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	cfg := config.GetGuildConfig(i.GuildID)

	resp := cmd.NewResponseBuilder(s, i)

	msgData := i.MessageComponentData()
	switch msgData.CustomID {
	case ConfigModifyXpRole:
		err := resp.IsEphemeral().
			SetMessage("Action à réaliser").
			SetComponents(component.New().Add(new(component.ActionRow).Add(
				new(component.StringSelect).SetPlaceholder("Action").
					AddOption(
						component.NewSelectOption("Ajouter", XpRoleAdd).
							SetDescription("Ajouter un rôle à XP").
							SetEmoji(&discordgo.ComponentEmoji{Name: "⬆️"}),
					).
					AddOption(
						component.NewSelectOption("Modifier", XpRoleEdit).
							SetDescription("Modifier un rôle à XP").
							SetEmoji(&discordgo.ComponentEmoji{Name: "📝"}),
					).
					AddOption(
						component.NewSelectOption("Supprimer", XpRoleDel).
							SetDescription("Supprimer un rôle à XP").
							SetEmoji(&discordgo.ComponentEmoji{Name: "❌"}),
					).
					SetCustomID(ConfigModify),
			))).Send()
		if err != nil {
			logger.Alert("config/guild.go - Sending config", err.Error())
		}
	case XpRoleAdd, XpRoleEdit:
		cID := XpRoleAddLevel
		if msgData.CustomID == XpRoleEdit {
			cID = XpRoleEditLevel
		}
		component.New().ForModal().Add(new(component.ActionRow).Add(
			new(component.TextInput).
				SetLabel("Niveau").
				SetPlaceholder("5").
				IsRequired().
				SetMinLength(0).
				SetMaxLength(5).
				SetStyle(discordgo.TextInputShort).
				SetCustomID(cID),
		))
		err := resp.IsModal().
			SetTitle("Role").
			SetComponents(component.New().ForModal().Add(new(component.ActionRow).Add(
				new(component.TextInput).
					SetLabel("Niveau").
					SetPlaceholder("5").
					IsRequired().
					SetMinLength(0).
					SetMaxLength(5).
					SetStyle(discordgo.TextInputShort).
					SetCustomID(cID),
			))).
			Send()
		if err != nil {
			logger.Alert("config/guild.go - Sending modal to add", err.Error())
		}
	case XpRoleAddRole:
		roleId := msgData.Values[0]
		for _, r := range cfg.XpRoles {
			if r.RoleID == roleId {
				err := resp.SetMessage("Le rôle est déjà présent dans la config").Send()
				if err != nil {
					logger.Alert("commands/config.go - Role already in config", err.Error())
				}
				return
			}
		}
		cfg.XpRoles = append(cfg.XpRoles, config.XpRole{
			XP:     configModifyMap[getKeyConfigRole(i)],
			RoleID: roleId,
		})
		err := cfg.Save()
		if err != nil {
			logger.Alert(
				"commands/config.go - Saving config",
				err.Error(),
				"guild_id", i.GuildID,
				"role_id", roleId,
				"type", "add",
			)
		}
		if err = resp.IsEphemeral().SetMessage("Rôle ajouté.").Send(); err != nil {
			logger.Alert("commands/config.go - Sending success", err.Error())
		}
	case XpRoleEditRole:
		roleId := msgData.Values[0]
		_, r := cfg.FindXpRole(roleId)
		if r == nil {
			err := resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				logger.Alert("commands/config.go - Role not found (edit)", err.Error())
			}
			return
		}
		r.XP = configModifyMap[getKeyConfigRole(i)]
		err := gokord.DB.Save(r).Error
		if err != nil {
			logger.Alert(
				"commands/config.go - Saving config",
				err.Error(),
				"guild_id", i.GuildID,
				"role_id", roleId,
				"type", "edit",
			)
		}
		if err = resp.IsEphemeral().SetMessage("Rôle modifié.").Send(); err != nil {
			logger.Alert("commands/config.go - Sending success", err.Error())
		}
	case XpRoleDel:
		err := resp.IsEphemeral().
			SetMessage("Rôle à supprimer").
			SetComponents(component.New().Add(new(component.ActionRow).Add(new(component.RoleSelect).SetCustomID(XpRoleDelRole)))).
			Send()
		if err != nil {
			logger.Alert("config/guild.go - Sending response to del", err.Error())
		}
	case XpRoleDelRole:
		roleId := msgData.Values[0]
		_, r := cfg.FindXpRole(roleId)
		if r == nil {
			err := resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
			if err != nil {
				logger.Alert("commands/config.go - Role not found (del)", err.Error())
			}
			return
		}
		err := gokord.DB.Delete(r).Error
		if err != nil {
			logger.Alert(
				"commands/config.go - Deleting entry",
				err.Error(),
				"guild_id", i.GuildID,
				"role_id", roleId,
				"type", "del",
			)
		}
		if err = resp.IsEphemeral().SetMessage("Rôle supprimé.").Send(); err != nil {
			logger.Alert("commands/config.go - Sending success", err.Error())
		}
	default:
		err := resp.SetMessage("Le type d'action n'est pas valide.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Invalid action type", err.Error())
		}
		return
	}
}

func ConfigXPModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	resp := cmd.NewResponseBuilder(s, i)

	modalData := i.ModalSubmitData()

	if modalData.CustomID != XpRoleAddLevel && modalData.CustomID != XpRoleEditLevel {
		return
	}

	input := modalData.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)

	k := getKeyConfigRole(i)
	in, err := strconv.Atoi(input.Value)
	if err != nil || in < 0 {
		if err = resp.IsEphemeral().
			SetMessage("Impossible de lire le nombre. Il doit s'agit d'un nombre entier positif.").
			Send(); err != nil {
			logger.Alert("command/config.go - Sending bad number", err.Error())
		}
		return
	}
	configModifyMap[k] = uint(in)
	go func(i *discordgo.InteractionCreate, k string) {
		time.Sleep(5 * time.Minute)
		delete(configModifyMap, k)
	}(i, k)

	cID := XpRoleAddRole
	resp.SetMessage("Rôle à ajouter")
	if modalData.CustomID == XpRoleEditLevel {
		cID = XpRoleEditLevel
		resp.SetMessage("Rôle à modifier")
	}

	err = resp.IsEphemeral().
		SetMessage("Rôle à supprimer").
		SetComponents(component.New().Add(new(component.ActionRow).Add(new(component.RoleSelect).SetCustomID(cID)))).
		Send()
	if err != nil {
		logger.Alert("config/guild.go - Sending response to add/edit", err.Error())
	}
}

func getKeyConfigRole(i *discordgo.InteractionCreate) string {
	return fmt.Sprintf("r:%s:%s", i.GuildID, i.User.ID)
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

func ConfigFallbackChannel(s *discordgo.Session, i *discordgo.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	salon, ok := optMap["channel"]
	if !ok {
		err := resp.SetMessage("Le salon n'a pas été renseigné.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Channel not set (fallback)", err.Error())
		}
		return
	}
	channel := salon.ChannelValue(s)
	if channel.Type != discordgo.ChannelTypeGuildText {
		err := resp.SetMessage("Le salon n'est pas un salon textuel.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Invalid channel type", err.Error())
		}
		return
	}
	cfg := config.GetGuildConfig(i.GuildID)
	cfg.FallbackChannel = channel.ID
	// save
	err := cfg.Save()
	if err != nil {
		logger.Alert(
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
		logger.Alert("commands/config.go - Channel saved message", err.Error())
	}
}

func ConfigPeriodBeforeReduce(s *discordgo.Session, i *discordgo.InteractionCreate, optMap cmd.OptionMap, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	// verify every args
	days, ok := optMap["days"]
	if !ok {
		err := resp.SetMessage("Le nombre de jours n'a pas été renseigné.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Days not set (fallback)", err.Error())
		}
		return
	}
	d := days.IntValue()
	if d < 30 {
		err := resp.SetMessage("Le nombre de jours est inférieur à 30.").Send()
		if err != nil {
			logger.Alert("commands/config.go - Days < 30 (fallback)", err.Error())
		}
		return
	}
	// save
	cfg := config.GetGuildConfig(i.GuildID)
	cfg.DaysXPRemains = uint(d)
	err := cfg.Save()
	if err != nil {
		logger.Alert(
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
		logger.Alert("commands/config.go - Days saved message", err.Error())
	}
}
