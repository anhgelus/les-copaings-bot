package rolereact

import (
	"fmt"
	"slices"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

func MessageModifyData(i *event.InteractionCreate, parameters *EditID) *interaction.ResponseData {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	if !ok {
		return &interaction.ResponseData{
			Flags: channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Cette modification est trop vieille et a été oubliée."},
			},
		}
	}
	var note string
	if message.Note != "" {
		note = message.Note
	} else {
		note = "*Pas de note*"
	}
	components := []component.Message{
		&component.TextDisplay{Content: "## Modifier un message de réaction"},
		&component.Separator{},
		&component.Section{
			Components: []component.Message{&component.TextDisplay{Content: note}},
			Accessory: &component.Button{
				Label:    "Modifier",
				Style:    component.ButtonStyleSecondary,
				CustomID: dynamicid.FormatCustomID(SetNote, *parameters),
			},
		},
		&component.Separator{},
	}
	for _, role := range message.Roles {
		var reaction string
		if role.Reaction != "" {
			reaction = FormatEmoji(role.Reaction)
		} else {
			reaction = ":no_entry_sign:"
		}
		var roleMention string
		if role.RoleID != "" {
			roleMention = fmt.Sprintf("<@&%s>", role.RoleID)
		} else {
			roleMention = "*Pas de rôle sélectionné*"
		}
		if role.CounterID == 0 {
			role.CounterID = roleCounter
			roleCounter++
		}
		components = append(components, &component.Section{
			Components: []component.Message{&component.TextDisplay{Content: fmt.Sprintf("%s %s", reaction, roleMention)}},
			Accessory: &component.Button{
				Label:    "Modifier",
				Style:    component.ButtonStyleSecondary,
				CustomID: dynamicid.FormatCustomID(OpenRole, EditIDWithRole{parameters.MessageEditID, role.CounterID}),
			},
		})
	}
	if len(message.Roles) == 0 {
		components = append(components, &component.TextDisplay{
			Content: "*Pas de rôles de réaction défini*",
		})
	}
	components = append(components, []component.Message{
		&component.ActionsRow{
			Components: []component.Message{
				&component.Button{
					Style:    component.ButtonStylePrimary,
					Label:    "Ajouter",
					CustomID: dynamicid.FormatCustomID(NewRole, EditID{MessageEditID: parameters.MessageEditID}),
					Disabled: len(message.Roles) >= 20,
				},
			},
		},
		&component.Separator{},
		&component.ActionsRow{
			Components: []component.Message{
				&component.Button{
					Label:    "Appliquer",
					Style:    component.ButtonStylePrimary,
					CustomID: dynamicid.FormatCustomID(ApplyMessage, EditID{MessageEditID: parameters.MessageEditID}),
				},
				&component.Button{
					Label:    "Réinitialiser",
					Style:    component.ButtonStyleDanger,
					CustomID: dynamicid.FormatCustomID(ResetMessage, *parameters),
				},
				&component.Button{
					Label: "Message",
					Style: component.ButtonStyleLink,
					URL:   fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, message.ChannelID, message.MessageID),
				},
			},
		}}...)
	responseData := &interaction.ResponseData{
		Flags: channel.MessageFlagsIsComponentsV2 | channel.MessageFlagsEphemeral,
		Components: []component.Component{
			&component.Container{
				Components: components,
			},
		},
	}
	return responseData
}

func MessageModifyRoleComponents(i *event.InteractionCreate, parameters *EditIDWithRole, emojiMessage string) []component.Message {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var role *config.RoleReact
	if ok {
		roleIndex := slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == parameters.RoleCounterID })
		if roleIndex != -1 {
			role = message.Roles[roleIndex]
		}
	}
	if !ok || role == nil {
		return []component.Message{
			&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
		}
	}
	disableBack := false
	var reactionDescription string
	var reactionButton component.Button
	if role.Reaction != "" {
		reactionDescription = fmt.Sprintf("**Réaction : ** %s", FormatEmoji(role.Reaction))
		reactionButton = component.Button{Label: "Modifier", Style: component.ButtonStyleSecondary}
	} else {
		reactionDescription = "*Aucune réaction pour le moment*"
		reactionButton = component.Button{Label: "Ajouter", Style: component.ButtonStylePrimary}
		disableBack = true
	}
	reactionButton.CustomID = dynamicid.FormatCustomID(SetRoleReaction, *parameters)
	defaultRoleValues := make([]component.SelectMenuDefaultValue, 0)
	if role.RoleID != "" {
		defaultRoleValues = append(defaultRoleValues, component.SelectMenuDefaultValue{
			Type: types.SelectMenuDefaultValueRole,
			ID:   role.RoleID,
		})
	}
	disableBack = disableBack || (role.RoleID == "")
	one := 1
	components := []component.Message{
		&component.TextDisplay{Content: "## Modifier un message de réaction"},
		&component.Separator{},
		&component.Section{
			Components: []component.Message{
				&component.TextDisplay{Content: reactionDescription},
			},
			Accessory: &reactionButton,
		},
	}
	if emojiMessage != "" {
		components = append(components, &component.TextDisplay{Content: "-# " + emojiMessage})
	}
	components = append(components,
		[]component.Message{
			&component.ActionsRow{Components: []component.Message{
				&component.SelectMenu{
					MenuType:  types.SelectMenuRole,
					CustomID:  dynamicid.FormatCustomID(SetRoleRoleID, *parameters),
					MinValues: &one, MaxValues: 1,
					Placeholder:   "Sélectionner un rôle",
					DefaultValues: defaultRoleValues,
				},
			}},
			&component.ActionsRow{Components: []component.Message{
				&component.Button{
					Style:    component.ButtonStyleDanger,
					Label:    "Supprimer",
					CustomID: dynamicid.FormatCustomID(DelRole, *parameters),
				},
			}},
			&component.Separator{},
			&component.ActionsRow{Components: []component.Message{
				&component.Button{
					Label:    "Retour",
					Style:    component.ButtonStyleSecondary,
					Disabled: disableBack,
					CustomID: dynamicid.FormatCustomID(OpenMessage, EditID{MessageEditID: parameters.MessageEditID}),
				},
				&component.Button{
					Label: "Message", Style: component.ButtonStyleLink,
					URL: fmt.Sprintf("https://discord.com/channels/%s/%s/%s", message.GuildID, message.ChannelID, message.MessageID),
				},
			}},
		}...)
	return []component.Message{&component.Container{
		Components: components,
	}}
}

func MessageModifyRoleData(i *event.InteractionCreate, parameters *EditIDWithRole, emojiMessage string) interaction.ResponseData {
	components := []component.Component{}
	for _, component := range MessageModifyRoleComponents(i, parameters, emojiMessage) {
		components = append(components, component)
	}
	return interaction.ResponseData{
		Flags:      channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
		Components: components,
	}
}
