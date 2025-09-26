package rolereact

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	oldGokord "github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/emoji"
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

var messageCounter uint = 1
var roleCounter uint = 1
var messageEdits map[uint]*config.RoleReactMessage = make(map[uint]*config.RoleReactMessage)

func MessageContent(message *config.RoleReactMessage) string {
	content := "## Réagis pour obtenir un rôle"
	if message.Note != "" {
		content = fmt.Sprintf("%s\n%s", content, message.Note)
	}
	for _, role := range message.Roles {
		if role.Reaction != "" && role.RoleID != "" {
			content += fmt.Sprintf("\n> -# %s <@&%s>", FormatEmoji(role.Reaction), role.RoleID)
		}
	}
	if len(message.Roles) == 0 {
		content += "\n*Pas de rôles pour le moment*"
	}
	return content
}

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
	responsedata := &interaction.ResponseData{
		Flags: channel.MessageFlagsIsComponentsV2 | channel.MessageFlagsEphemeral,
		Components: []component.Component{
			&component.Container{
				Components: components,
			},
		},
	}
	return responsedata
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

func HandleCommand(
	s bot.Session,
	i *event.InteractionCreate,
	o cmd.OptionMap,
	resp *cmd.ResponseBuilder,
) {
	s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseDeferredMessageUpdate,
		Data: &interaction.ResponseData{Flags: channel.MessageFlagsEphemeral},
	})
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
			s.Logger().Error("Unable to send reset message message", "error", err)
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
			s.Logger().Error("Unable to send reset message message", "error", err)
		}
	} else {
		err := s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseDeferredChannelMessageWithSource,
			Data: &interaction.ResponseData{Flags: channel.MessageFlagsEphemeral},
		})
		if err != nil {
			s.Logger().Error("Unable to defer interaction", "error", err)
			return
		}
		messageContent := MessageContent(message)
		_, err = s.ChannelAPI().MessageEditComplex(
			&channel.MessageEdit{
				Content:         &messageContent,
				AllowedMentions: &channel.MessageAllowedMentions{},
				Channel:         message.ChannelID,
				ID:              message.MessageID,
			},
		)
		if err == nil {
			for _, role := range message.Roles {
				if role.Reaction != "" && role.RoleID != "" && err == nil {
					err = s.ChannelAPI().MessageReactionAdd(
						message.ChannelID,
						message.MessageID,
						role.Reaction,
					)
				}
			}
		}
		var content string
		if err != nil {
			content = "Impossible de mettre à jour le message."
			s.Logger().Error("Unable to apply rolereaction message changes", "error", err)
			err = nil
		} else {
			content = "Message mis à jour avec succès."
			cfg := GetGuildConfigPreloaded(i.GuildID)
			messageIndex := slices.IndexFunc(cfg.RrMessages, func(m config.RoleReactMessage) bool { return m.ID == message.ID })
			if messageIndex != -1 {
				oldMessage := cfg.RrMessages[messageIndex]
				// cfg.RrMessages[messageIndex] = *message
				roles := make(map[uint]config.RoleReact, len(message.Roles))
				for _, role := range message.Roles {
					roles[role.ID] = *role
				}
				for _, role := range oldMessage.Roles {
					_, ok := roles[role.ID]
					if !ok {
						oldGokord.DB.Delete(role)
					}
				}
				// cfg.Save()
				cfg.RrMessages[messageIndex] = *message
				oldGokord.DB.Save(cfg.RrMessages[messageIndex])
				for _, role := range cfg.RrMessages[messageIndex].Roles {
					oldGokord.DB.Save(role)
				}
			}
		}
		_, err = s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			s.Logger().Error("Unable to send apply rolereaction message changes", "error", err)
		}
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	emojiChann := make(chan emoji.Emoji)

	cancelHandler := s.EventManager().AddHandler(func(s bot.Session, e *event.MessageReactionAdd) {
		if e.MessageID == message.MessageID && e.UserID == i.Member.User.ID {
			emojiChann <- e.Emoji
		}
	})
	defer cancelHandler()

	select {
	case emoji := <-emojiChann:
		emojiName := emoji.APIName()
		err := s.ChannelAPI().MessageReactionAdd(message.ChannelID, message.MessageID, emojiName)
		if err != nil {
			responseData := MessageModifyRoleData(i, parameters, "La réaction n'est pas utilisable. Cela peut être résolu en l'ajoutant à ce serveur")
			err = s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
				Type: types.InteractionResponseUpdateMessage,
				Data: &responseData,
			})
			if err != nil {
				s.Logger().Error("Unable to edit original response", "error", err)
			}
			return
		}
		s.ChannelAPI().MessageReactionRemove(message.ChannelID, message.MessageID, emojiName, i.Member.User.ID)
		role.Reaction = emojiName
		components := MessageModifyRoleComponents(i, parameters, "")
		_, err = s.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Flags:      channel.MessageFlagsIsComponentsV2 | channel.MessageFlagsEphemeral,
			Components: &components,
		})
		if err != nil {
			s.Logger().Error("Unable to edit original response", "error", err)
		}
	case <-ctx.Done():
		responseData := MessageModifyRoleData(i, parameters, "Le temps d'attente a été dépassé")
		s.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &responseData,
		})
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

func GetMessageFromEditID(i *event.InteractionCreate, editID uint) (*config.RoleReactMessage, bool) {
	cfg := config.GetGuildConfig(i.GuildID)
	m, ok := messageEdits[editID]
	if !ok || m.GuildConfigID != cfg.ID {
		return &config.RoleReactMessage{}, false
	}
	return m, true
}

func GetGuildConfigPreloaded(guildID string) *config.GuildConfig {
	cfg := config.GuildConfig{GuildID: guildID}
	// err := oldGokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").Preload("RrMessages.Roles").FirstOrCreate(cfg).Error
	err := oldGokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("RrMessages.Roles").FirstOrCreate(&cfg).Error
	if err != nil {
		panic(err)
	}
	return &cfg
}

func FormatEmoji(apiName string) string {
	if strings.Contains(apiName, ":") {
		return fmt.Sprintf("<:%s>", apiName)
	} else {
		return apiName
	}
}
