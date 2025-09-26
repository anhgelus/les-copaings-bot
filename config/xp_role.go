package config

import (
	"fmt"
	"slices"
	"strconv"

	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
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
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	_ *cmd.ResponseBuilder,
) {
	cfg := GetGuildConfig(i.GuildID)
	container := component.Container{
		Components: []component.Message{
			&component.TextDisplay{Content: "## Configuration / Rôles de niveaux"},
			&component.TextDisplay{Content: "Ces rôles seront donnés et retirés en fonction du niveau de chacun"},
			&component.Separator{},
		},
	}
	slices.SortFunc(cfg.XpRoles, func(xp1, xp2 XpRole) int {
		return int(xp2.XP) - int(xp1.XP)
	})
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
	err := s.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		s.LogError(err, "Sending config")
	}
}

func HandleXpRoleNew(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	_ *cmd.ResponseBuilder,
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
	err := s.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		s.LogError(err, "Sending modal to add")
	}
}

func HandleXpRoleEdit(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	parameters []string, resp *cmd.ResponseBuilder,
) {
	config := GetGuildConfig(i.GuildID)
	id, err := getRoleLevelID(parameters)
	if err != nil {
		s.LogError(err, "Reading dynamic CustomID")
		return
	}
	_, role := config.FindXpRoleID(id)
	if role == nil {
		HandleXpRole(s, i, &interaction.MessageComponentData{}, resp)
		return
	}

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

	err = s.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		s.LogError(err, "Sending xp_role config")
	}
}

func HandleXpRoleEditRole(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters []string, resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		s.LogError(err, "Reading dynamic CustomID")
		return
	}
	role := data.Values[0]
	cfg := GetGuildConfig(i.GuildID)
	_, xpRole := cfg.FindXpRoleID(id)
	if xpRole == nil {
		err = s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de modifier le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			s.LogError(err, "Sending unable to get role message")
		}
		return
	}
	xpRole.RoleID = role
	err = gokord.DB.Save(xpRole).Error
	if err != nil {
		s.LogError(err, "Saving config guild_id %s, id %d, type add", i.GuildID, id)
	}
	HandleXpRoleEdit(s, i, &interaction.MessageComponentData{}, parameters, resp)
}

func HandleXpRoleEditLevelStart(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	parameters []string,
	_ *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		s.LogError(err, "Reading dynamic CustomID")
		return
	}
	cfg := GetGuildConfig(i.GuildID)
	_, xpRole := cfg.FindXpRoleID(id)
	if xpRole == nil {
		err = s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de trouver le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			s.LogError(err, "Sending Unable to get role message")
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
						Value:       strconv.FormatUint(uint64(exp.Level(xpRole.XP)), 10),
					},
				},
			},
		},
	}
	err = s.InteractionAPI().Respond(i.Interaction, response)
	if err != nil {
		s.LogError(err, "Sending Edit level modal")
	}
}

func HandleXpRoleEditLevel(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.ModalSubmitData,
	parameters []string,
	resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(parameters)
	if err != nil {
		s.LogError(err, "Reading dynamic CustomID")
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
			s.LogError(err, "Sending bad number warning message")
		}
		return
	}
	xp := exp.LevelXP(uint(level))

	cfg := GetGuildConfig(i.GuildID)
	_, xpRole := cfg.FindXpRoleID(id)
	if xpRole == nil {
		err = s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Flags:   channel.MessageFlagsEphemeral,
				Content: "Impossible de modifier le rôle. Peut-être a-t-il été supprimé ?",
			},
		})
		if err != nil {
			s.LogError(err, "Sending unable to modify role message")
		}
		return
	}
	xpRole.XP = xp
	err = gokord.DB.Save(xpRole).Error
	if err != nil {
		s.LogError(err, "Saving config guild_id %s, id %d, type add", i.GuildID, id)
	}
	HandleXpRoleEdit(s, i, &interaction.MessageComponentData{}, parameters, resp)
}

func HandleXpRoleDel(
	s bot.Session,
	i *event.InteractionCreate,
	_ *interaction.MessageComponentData,
	dynamicValues []string,
	resp *cmd.ResponseBuilder,
) {
	id, err := getRoleLevelID(dynamicValues)
	if err != nil {
		s.LogError(err, "reading dynamic CustomID")
		return
	}
	cfg := GetGuildConfig(i.GuildID)
	_, role := cfg.FindXpRoleID(id)
	if role == nil {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{
				Content: "Rôle introuvable. Peut-être a-t-il déjà été supprimé ?",
				Flags:   channel.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			s.LogError(err, "Sending role not found message")
		}
		return
	}
	err = gokord.DB.Delete(role).Error
	if err != nil {
		s.LogError(err, "Deleting entry guild_id %s, id %d, type del", i.GuildID, id)
	}

	HandleXpRole(s, i, &interaction.MessageComponentData{}, resp)
}

func HandleXpRoleAdd(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.ModalSubmitData,
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
			s.LogError(err, "sending bad number warning message")
		}
		return
	}
	xp := exp.LevelXP(uint(in))

	roleId := data.Components[1].(*component.Label).Component.(*component.SelectMenu).Values[0]

	cfg := GetGuildConfig(i.GuildID)
	cfg.XpRoles = append(cfg.XpRoles, XpRole{
		XP:     xp,
		RoleID: roleId,
	})
	err = cfg.Save()
	if err != nil {
		s.LogError(err, "saving config for role %s in %s", roleId, i.GuildID)
		return
	}

	HandleXpRole(s, i, &interaction.MessageComponentData{}, resp)
}

func getRoleLevelID(dynamic []string) (uint, error) {
	id64, err := strconv.ParseUint(dynamic[0], 10, 0)
	if err != nil {
		return 0, err
	}

	return uint(id64), nil
}
