package config

import (
	"fmt"
	"slices"
	"strconv"

	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/interaction"
)

type XpRole struct {
	ID            uint `gorm:"primarykey"`
	XP            uint
	RoleID        string
	GuildConfigID uint
}

const (
	ModifyXpRole                = "xp_role"
	XpRoleNew                   = "xp_role_add"
	XpRoleAdd                   = "xp_role_add_level"
	XpRoleEditPattern           = `^xp_role_edit_(\d+)$`
	XpRoleEditLevelPattern      = `^xp_role_edit_level_(\d+)$`
	XpRoleEditLevelStartPattern = `^xp_role_edit_level_start_(\d+)$`
	XpRoleEditRolePattern       = `^xp_role_edit_role_(\d+)$`
	XpRoleDel                   = `^xp_role_del_(\d+)$`
)

func HandleXpRole(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ interaction.MessageComponentData,
	resp *cmd.ResponseBuilder,
) {
	cfg := GetGuildConfig(i.GuildID)
	container := component.Container{
		Components: []component.Message{
			&component.TextDisplay{Content: "## Configuration / Rôles de niveaux"},
			&component.Separator{},
		},
	}
	for _, r := range cfg.XpRoles {
		container.Components = append(container.Components, &component.Section{
			Components: []component.Message{
				&component.TextDisplay{
					Content: fmt.Sprintf("<@&%s> - Niveau %d", r.RoleID, exp.Level(r.XP)),
				},
			},
			Accessory: &component.Button{
				CustomID: fmt.Sprintf("xp_role_edit_%d", r.ID),
				Style:    component.ButtonStyleSecondary,
				Label:    "Modifier",
			},
		})
	}
	container.Components = append(container.Components,
		&component.ActionsRow{
			Components: []component.Message{
				&component.Button{
					CustomID: XpRoleNew,
					Style:    component.ButtonStylePrimary,
					Label:    "Nouveau rôle",
				},
			},
		},
		&component.Separator{},
		&component.ActionsRow{
			Components: []component.Message{
				&component.Button{CustomID: "config", Style: component.ButtonStyleSecondary, Label: "Retour"},
			},
		},
	)

	response := &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &interaction.ResponseData{
			Components: []component.Component{&container},
			Flags:      channel.MessageFlagsIsComponentsV2,
		},
	}
	err := session.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		session.LogError(err, "Sending config")
	}
}

func HandleXpRoleNew(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.MessageComponentData,
	resp *cmd.ResponseBuilder,
) {
	one := 1
	response := &interaction.Response{
		Type: types.InteractionResponseModal,
		Data: &interaction.ResponseData{
			Title:    "Nouveau rôle de niveau",
			CustomID: XpRoleAdd,
			Components: []component.Component{
				&component.Label{
					Label: "Niveau",
					Component: &component.TextInput{
						CustomID:    "level",
						Style:       component.TextInputShort,
						Placeholder: "5",
						MinLength:   1,
						MaxLength:   5,
						Required:    true,
					},
				},
				&component.Label{
					Label: "Rôle",
					Component: &component.SelectMenu{
						MenuType:  types.SelectMenuRole,
						CustomID:  "role",
						MinValues: &one,
						MaxValues: one,
					},
				},
			},
		},
	}
	err := session.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		session.LogError(err, "Sending modal to add")
	}
}

func HandleXpRoleEdit(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.MessageComponentData,
	parameters []string, resp *cmd.ResponseBuilder,
) {
	config := GetGuildConfig(i.GuildID)
	id, err := getRoleLevelID(parameters)
	if err != nil {
		session.LogError(err, "Reading dynamic CustomID")
		return
	}
	roleIndex := slices.IndexFunc(config.XpRoles, func(role XpRole) bool { return role.ID == id })
	if roleIndex == -1 {
		return
	}
	role := config.XpRoles[roleIndex]

	roleSelect := &component.SelectMenu{
		MenuType: types.SelectMenuRole,
		CustomID: fmt.Sprintf("xp_role_edit_role_%d", id),
		DefaultValues: []component.SelectMenuDefaultValue{
			{ID: role.RoleID, Type: types.SelectMenuDefaultValueRole},
		},
	}

	container := &component.Container{
		Components: []component.Message{
			&component.TextDisplay{Content: "## Configuration / Rôles de niveaux"},
			&component.Separator{},
			&component.Section{
				Components: []component.Message{
					&component.TextDisplay{Content: fmt.Sprintf("Niveau **%d**", exp.Level(role.XP))},
				},
				Accessory: &component.Button{
					CustomID: fmt.Sprintf("xp_role_edit_level_start_%d", id),
					Style:    component.ButtonStyleSecondary,
					Label:    "Modifier",
				},
			},
			&component.ActionsRow{Components: []component.Message{roleSelect}},
			&component.ActionsRow{Components: []component.Message{
				&component.Button{CustomID: fmt.Sprintf("xp_role_del_%d", id), Style: component.ButtonStyleDanger, Label: "Supprimer"},
			}},
			&component.Separator{},
			&component.ActionsRow{Components: []component.Message{
				&component.Button{Label: "Retour", CustomID: ModifyXpRole, Style: component.ButtonStyleSecondary},
			}},
		},
	}

	response := &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &interaction.ResponseData{
			Components: []component.Component{container},
			Flags:      channel.MessageFlagsIsComponentsV2,
		},
	}

	err = session.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		session.LogError(err, "Sending xp_role config")
	}
}

func HandleXpRoleEditRole(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.MessageComponentData,
	parameters []string, resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		session.LogError(err, "Reading dynamic CustomID")
		return
	}
	role := data.Values[0]
	cfg := GetGuildConfig(i.GuildID)
	for _, xpRole := range cfg.XpRoles {
		if xpRole.RoleID == role {
			err = session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
				Type: types.InteractionResponseChannelMessageWithSource,
				Data: &interaction.ResponseData{
					Flags:           channel.MessageFlagsEphemeral,
					AllowedMentions: &channel.MessageAllowedMentions{},
					Content:         fmt.Sprintf("Un autre niveau avec le rôle <@&%s> est déjà existant.", role),
				},
			})
			if err != nil {
				session.LogError(err, "Sending unable to Already existing role message")
			}
			return
		}
	}
	index, xprole := cfg.FindXpRoleID(id)
	if index == 0 {
		err = session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de modifier le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			session.LogError(err, "Sending unable to get role message")
		}
		return
	}
	xprole.RoleID = role
	err = gokord.DB.Save(xprole).Error
	if err != nil {
		session.LogError(err, "Saving config guild_id %s, id %d, type add", i.GuildID, id)
	}
	HandleXpRoleEdit(session, i, interaction.MessageComponentData{}, parameters, resp)
}

func HandleXpRoleEditLevelStart(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.MessageComponentData,
	parameters []string,
	resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		session.LogError(err, "Reading dynamic CustomID")
		return
	}
	cfg := GetGuildConfig(i.GuildID)
	_, role := cfg.FindXpRoleID(id)
	if role == nil {
		err = session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de trouver le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			session.LogError(err, "Sending Unable to get role message")
		}
		return
	}
	response := &interaction.Response{
		Type: types.InteractionResponseModal,
		Data: &interaction.ResponseData{
			Title:    "Modification du niveau lié au rôle",
			CustomID: fmt.Sprintf("xp_role_edit_level_%d", id),
			Components: []component.Component{
				&component.Label{
					Label: "Nouveau niveau",
					Component: &component.TextInput{
						Style:       component.TextInputShort,
						Required:    true,
						CustomID:    "level",
						MinLength:   1,
						MaxLength:   5,
						Placeholder: "5",
						Value:       strconv.FormatUint(uint64(exp.Level(role.XP)), 10),
					},
				},
			},
		},
	}
	err = session.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		session.LogError(err, "Sending Edit level modal")
	}
}

func HandleXpRoleEditLevel(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.ModalSubmitData,
	parameters []string,
	resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		session.LogError(err, "Reading dynamic CustomID")
		return
	}

	fmt.Printf("Alors?... %#v", data.Components)
	levelInput := data.Components[0].(*component.Label).Component.(*component.TextInput)
	level, err := strconv.Atoi(levelInput.Value)
	if err != nil || level < 0 {
		err = resp.IsEphemeral().
			SetMessage(
				fmt.Sprintf("Le niveau doit être un nombre entier positif.\n-# Trouvé : %s", levelInput.Value),
			).
			Send()
		if err != nil {
			session.LogError(err, "Sending bad number warning message")
		}
		return
	}
	xp := exp.LevelXP(uint(level))

	cfg := GetGuildConfig(i.GuildID)
	for _, xpRole := range cfg.XpRoles {
		if xpRole.XP == xp {
			err = session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
				Type: types.InteractionResponseChannelMessageWithSource,
				Data: &interaction.ResponseData{
					Flags:   channel.MessageFlagsEphemeral,
					Content: fmt.Sprintf("Un autre rôle est déjà lié au niveau %d.", level),
				},
			})
			if err != nil {
				session.LogError(err, "Sending unable to Already existing level message")
			}
			return
		}
	}
	index, xprole := cfg.FindXpRoleID(id)
	if index == -1 {
		err = session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de modifier le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			session.LogError(err, "Sending unable to modify role message")
		}
		return
	}
	xprole.XP = xp
	err = gokord.DB.Save(xprole).Error
	if err != nil {
		session.LogError(err, "Saving config guild_id %s, id %d, type add", i.GuildID, id)
	}
	HandleXpRoleEdit(session, i, interaction.MessageComponentData{}, parameters, resp)
}

func HandleXpRoleDel(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	_ interaction.MessageComponentData,
	dynamic_values []string,
	resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(dynamic_values)
	if err != nil {
		session.LogError(err, "reading dynamic CustomID")
		return
	}
	cfg := GetGuildConfig(i.GuildID)
	_, role := cfg.FindXpRoleID(id)
	if role == nil {
		err := session.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Content: "Rôle introuvable. Peut-être a-t-il déjà été supprimé ?",
				Flags:   channel.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			session.LogError(err, "Sending role not found message")
		}
		return
	}
	err = gokord.DB.Delete(role).Error
	if err != nil {
		session.LogError(err, "Deleting entry guild_id %s, id %d, type del", i.GuildID, id)
	}

	HandleXpRole(session, i, interaction.MessageComponentData{}, resp)
}

func HandleXpRoleAdd(
	session *discordgo.Session,
	i *discordgo.InteractionCreate,
	data interaction.ModalSubmitData,
	resp *cmd.ResponseBuilder,
) {
	levelInput := data.Components[0].(*component.Label).Component.(*component.TextInput)

	in, err := strconv.Atoi(levelInput.Value)
	if err != nil || in < 0 {
		err = resp.IsEphemeral().
			SetMessage(
				fmt.Sprintf("Le niveau doit être un nombre entier positif.\n-# Trouvé : %s", levelInput.Value),
			).
			Send()
		if err != nil {
			session.LogError(err, "sending bad number warning message")
		}
		return
	}
	xp := exp.LevelXP(uint(in))

	roleId := data.Components[1].(*component.Label).Component.(*component.SelectMenu).Values[0]

	cfg := GetGuildConfig(i.GuildID)
	for _, r := range cfg.XpRoles {
		if r.RoleID == roleId {
			err := resp.IsEphemeral().SetMessage(fmt.Sprintf("Le rôle <@&%s> est déjà lié au niveau %d.", r.RoleID, exp.Level(r.XP))).Send()
			if err != nil {
				session.LogError(err, "sending role already in config")
			}
			return
		} else if r.XP == xp {
			err := resp.IsEphemeral().SetMessage(fmt.Sprintf("Le niveau %d est déjà lié au rôle <@&%s>.", in, r.RoleID)).Send()
			if err != nil {
				session.LogError(err, "sending role already in config")
			}
			return
		}
	}
	cfg.XpRoles = append(cfg.XpRoles, XpRole{
		XP:     xp,
		RoleID: roleId,
	})
	err = cfg.Save()
	if err != nil {
		session.LogError(err, "saving config for role %s in %s", roleId, i.GuildID)
		return
	}

	HandleXpRole(session, i, interaction.MessageComponentData{}, resp)
}

func getRoleLevelID(dynamic []string) (uint, error) {
	id64, err := strconv.ParseUint(dynamic[0], 10, 0)
	if err != nil {
		return 0, err
	}

	return uint(id64), nil
}
