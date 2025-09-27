package rolereact

import (
	"fmt"
	"slices"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

const (
	OpenMessage     = "rolereact_message"
	ResetMessage    = "rolereact_reset_message"
	ApplyMessage    = "rolereact_apply_message"
	SetNote         = "rolereact_set_note"
	NewRole         = "rolereact_new_role"
	AddRole         = "rolereact_add_role"
	OpenRole        = "rolereact_open_role"
	SetRoleRoleID   = "rolereact_set_role_roleid"
	SetRoleReaction = "rolereact_set_role_reaction"
	DelRole         = "rolereact_del_role"
)

type EditID struct {
	MessageEditID uint
}

type EditIDWithRole struct {
	MessageEditID uint
	RoleCounterID uint
}

var (
	messageCounter uint                              = 1
	roleCounter    uint                              = 1
	messageEdits   map[uint]*config.RoleReactMessage = make(map[uint]*config.RoleReactMessage)
)

func HandleCommand(
	s bot.Session,
	i *event.InteractionCreate,
	o cmd.OptionMap,
	resp *cmd.ResponseBuilder,
) {
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseDeferredMessageUpdate,
		Data: &interaction.ResponseData{Flags: channel.MessageFlagsEphemeral},
	})
	if err != nil {
		s.Logger().Error("unable to defer interaction", "error", err)
		return
	}
	c := o["salon"]
	var channelID string
	if c != nil {
		channelID = c.Value.(string)
	} else {
		channelID = i.ChannelID
	}

	message := config.RoleReactMessage{
		ChannelID: channelID,
		GuildID:   i.GuildID,
	}
	messageContent := MessageContent(&message)
	m, err := s.ChannelAPI().MessageSendComplex(
		channelID, &channel.MessageSend{
			Content:         messageContent,
			AllowedMentions: &channel.MessageAllowedMentions{},
		},
	)
	if err != nil {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource,
			Data: &interaction.ResponseData{Content: fmt.Sprintf("Error: %s", err.Error())},
		})
		if err != nil {
			s.Logger().Error("Unable to send message", "error", err)
		}
		return
	}
	message.MessageID = m.ID
	cfg := GetGuildConfigPreloaded(i.GuildID)
	cfg.RrMessages = append(cfg.RrMessages, message)
	err = cfg.Save()
	if err != nil {
		s.Logger().Error("Unable to save rolereact message in database", "error", err)
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseChannelMessageWithSource | types.InteractionResponseDeferredChannelMessageWithSource,
			Data: &interaction.ResponseData{Content: "Unable to save message in database. Please retry later."},
		})
		if err != nil {
			s.Logger().Error("Unable to send message", "error", err)
		}
		return
	}

	messageEdits[messageCounter] = &cfg.RrMessages[len(cfg.RrMessages)-1]
	editID := messageCounter
	messageCounter++

	err = s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseChannelMessageWithSource,
		Data: MessageModifyData(i, &EditID{MessageEditID: editID}),
	})
	if err != nil {
		s.Logger().Error("Unable to send edit rolereact message", "error", err)
	}
}

func HandleModifyCommand(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.CommandInteractionData,
	resp *cmd.ResponseBuilder,
) {
	messageId := data.TargetID
	cfg := GetGuildConfigPreloaded(i.GuildID)
	var target *config.RoleReactMessage
	var targetEditID uint
	for editID, message := range messageEdits {
		if message.MessageID == messageId {
			targetEditID = editID
			target = message
		}
	}
	if targetEditID == 0 {
		for _, message := range cfg.RrMessages {
			if message.MessageID == messageId {
				target = &message
			}
		}
		if target == nil {
			err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
				Type: types.InteractionResponseChannelMessageWithSource,
				Data: &interaction.ResponseData{
					Flags:   channel.MessageFlagsEphemeral,
					Content: "Le message sélectionné n'est pas un message de rôles de réaction.",
				},
			})
			if err != nil {
				s.Logger().Error("Unable to send rolereact message not found", "error", err)
			}
			return
		}
		messageEdits[messageCounter] = target
		targetEditID = messageCounter
		messageCounter++
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseChannelMessageWithSource,
		Data: MessageModifyData(i, &EditID{MessageEditID: targetEditID}),
	})
	if err != nil {
		s.Logger().Error("Unable to send modify rolereact message", "error", err)
	}
}

func HandleModifyComponent(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: MessageModifyData(i, parameters),
	})
	if err != nil {
		s.Logger().Error("Unable to send modify rolereact message", "error", err)
	}
}

func HandleResetMessage(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		cfg := GetGuildConfigPreloaded(i.GuildID)
		for _, m := range cfg.RrMessages {
			if m.ID == message.ID {
				messageEdits[parameters.MessageEditID] = &m
			}
		}
		responseData = *MessageModifyData(i, parameters)
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	})
	if err != nil {
		s.Logger().Error("Unable to send reset message message", "error", err)
	}
}

func HandleStartSetNote(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	if !ok {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		})
		if err != nil {
			s.Logger().Error("Unable to send message edit not found message", "error", err)
		}
		return
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseModal,
		Data: &interaction.ResponseData{
			Title:    "Changer la description",
			CustomID: dynamicid.FormatCustomID(SetNote, *parameters),
			Components: []component.Component{
				&component.Label{
					Label:       "Nouvelle description",
					Description: "Description affichée sur votre message de réaction",
					Component: &component.TextInput{
						Style:     component.TextInputParagraph,
						MaxLength: 2000,
						CustomID:  "note",
						Value:     message.Note,
					},
				},
			},
		},
	})
	if err != nil {
		s.Logger().Error("Unable to send edit note modal", "error", err)
	}
}

func HandleSetNote(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.ModalSubmitData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	if !ok {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		})
		if err != nil {
			s.Logger().Error("unable to send set note error message", "error", err)
		}
		return
	}
	message.Note = data.Components[0].(*component.Label).Component.(*component.TextInput).Value
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: MessageModifyData(i, parameters),
	})
	if err != nil {
		s.Logger().Error("Unable to send updated note message", "error", err)
	}
}

func HandleApplyMessage(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &responseData,
		})
		if err != nil {
			s.Logger().Error("unable to send apply message error message", "error", err)
		}
		return
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseDeferredChannelMessageWithSource,
		Data: &interaction.ResponseData{Flags: channel.MessageFlagsEphemeral},
	})
	if err != nil {
		s.Logger().Error("Unable to defer interaction", "error", err)
		return
	}
	m := ApplyMessageChange(s, i, message)
	_, err = s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
		Content: &m,
	})
	if err != nil {
		s.Logger().Error("Unable to send apply rolereaction message changes", "error", err)
	}
}

func HandleNewRole(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditID,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		message.Roles = append(message.Roles, &config.RoleReact{CounterID: roleCounter})
		responseData = MessageModifyRoleData(i, &EditIDWithRole{MessageEditID: parameters.MessageEditID, RoleCounterID: roleCounter}, "")
		roleCounter++
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	})
	if err != nil {
		s.Logger().Error("Unable to send modify reaction role message", "error", err)
	}
}

func HandleOpenRole(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditIDWithRole,
	resp *cmd.ResponseBuilder,
) {
	_, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		responseData = MessageModifyRoleData(i, parameters, "")
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	})
	if err != nil {
		s.Logger().Error("Unable to send open reaction role message", "error", err)
	}
}

func HandleSetRole(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditIDWithRole,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var responseData interaction.ResponseData
	var role *config.RoleReact
	if ok {
		roleIndex := slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == parameters.RoleCounterID })
		if roleIndex != -1 {
			role = message.Roles[roleIndex]
		}
	}
	if !ok || role == nil {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		role.RoleID = data.Values[0]
		responseData = MessageModifyRoleData(i, parameters, "")
	}
	err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	})
	if err != nil {
		s.Logger().Error("Unable to send open reaction role message", "error", err)
	}
}

func HandleSetReaction(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditIDWithRole,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	var role *config.RoleReact
	if ok {
		roleIndex := slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == parameters.RoleCounterID })
		if roleIndex != -1 {
			role = message.Roles[roleIndex]
		}
	}
	if !ok || role == nil {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		})
		if err != nil {
			s.Logger().Error("Unable to send open reaction role message", "error", err)
		}
		return
	}
	responseData := MessageModifyRoleData(i, parameters, "Ajoute la réaction que tu veux choisir au message de rôle de réaction (tu peux y accéder avec le bouton ci-dessous)")
	s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	})
	emojiName, ok := WaitForEmoji(s, i.Member.User.ID, message.MessageID)
	if !ok {
		editResponseComponents := MessageModifyRoleComponents(i, parameters, "Le temps d'attente a été dépassé")
		_, err := s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Components: &editResponseComponents,
		})
		if err != nil {
			s.Logger().Error("unable to send timed out reaction message", "error", err)
		}
		return
	}

	err := s.ChannelAPI().MessageReactionAdd(message.ChannelID, message.MessageID, emojiName)
	if err != nil {
		editResponseComponents := MessageModifyRoleComponents(i, parameters, "La réaction n'est pas utilisable. Cela peut être résolu en l'ajoutant à ce serveur")
		_, err := s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Components: &editResponseComponents,
		})
		if err != nil {
			s.Logger().Error("unable to send unusable reaction message", "error", err)
		}
		return
	}
	err = s.ChannelAPI().MessageReactionRemove(message.ChannelID, message.MessageID, emojiName, i.Member.User.ID)
	if err != nil {
		s.Logger().Warn("unable to remove author reaction from message", "error", err)
	}
	role.Reaction = emojiName
	components := MessageModifyRoleComponents(i, parameters, "")
	_, err = s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
		Flags:      channel.MessageFlagsIsComponentsV2 | channel.MessageFlagsEphemeral,
		Components: &components,
	})
	if err != nil {
		s.Logger().Error("Unable to edit original response", "error", err)
	}
}

func HandleDelRole(
	s bot.Session,
	i *event.InteractionCreate,
	data *interaction.MessageComponentData,
	parameters *EditIDWithRole,
	resp *cmd.ResponseBuilder,
) {
	message, ok := GetMessageFromEditID(i, parameters.MessageEditID)
	roleIndex := -1
	if ok {
		roleIndex = slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == parameters.RoleCounterID })
	}
	if !ok || roleIndex == -1 {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		})
		if err != nil {
			s.Logger().Error("Unable to send open reaction role message", "error", err)
		}
		return
	}
	message.Roles = append(message.Roles[:roleIndex],
		message.Roles[roleIndex+1:]...,
	)
	err := s.InteractionAPI().Respond(i.Interaction,
		&interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: MessageModifyData(i, &EditID{MessageEditID: parameters.MessageEditID}),
		})
	if err != nil {
		s.Logger().Error("Unable to send modify message message", "error", err)
	}
}
