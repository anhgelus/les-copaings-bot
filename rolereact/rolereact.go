package rolereact

import (
	"context"
	"slices"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
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

func HandleCommand(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseDeferredChannelMessageWithSource,
		Data: &interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral,
		},
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to defer interaction", "error", err)
		return
	}
	o := i.OptionMap()
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
	m, err := dg.ChannelAPI().MessageSendComplex(channelID, &channel.MessageSend{
		Content:         messageContent,
		AllowedMentions: &channel.MessageAllowedMentions{},
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("sending message", "error", err)
		_, err := dg.InteractionAPI().FollowupMessageCreate(i.Interaction, false, &channel.WebhookParams{
			Flags:      channel.MessageFlagsIsComponentsV2,
			Components: []component.Message{&component.TextDisplay{Content: "Impossible d'envoyer le message"}},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send message", "error", err)
		}
		return
	}
	message.MessageID = m.ID
	cfg := GetGuildConfigPreloaded(ctx, i.GuildID)
	cfg.RrMessages = append(cfg.RrMessages, message)
	err = cfg.Save(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to save rolereact message in database", "error", err)
		_, err := dg.InteractionAPI().FollowupMessageCreate(i.Interaction, false, &channel.WebhookParams{
			Flags: channel.MessageFlagsIsComponentsV2,
			Components: []component.Message{
				&component.TextDisplay{Content: "Impossible d'enregistrer le message dans la base de données. Merci de contacter l'administrateur du bot."},
			},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send message", "error", err)
		}
		return
	}

	messageEdits[messageCounter] = &cfg.RrMessages[len(cfg.RrMessages)-1]
	editID := messageCounter
	messageCounter++

	components := MessageModifyComponents(ctx, i.Interaction, &EditID{MessageEditID: editID})
	// Using this function here is deprecated as per discord documentation.
	// That said, using the recommended EditResponse endpoint does not allow the use of
	// version 2 components...
	_, err = dg.InteractionAPI().FollowupMessageCreate(i.Interaction, false, &channel.WebhookParams{
		Flags:      channel.MessageFlagsIsComponentsV2,
		Components: components,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send edit rolereact message", "error", err)
	}
}

func HandleModifyCommand(ctx context.Context, dg bot.Session, i *interaction.ApplicationCommand) {
	logger := bot.Logger(ctx)
	messageId := i.Data.TargetID
	cfg := GetGuildConfigPreloaded(ctx, i.GuildID)
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
			err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
				Type: types.InteractionResponseChannelMessageWithSource,
				Data: &interaction.ResponseData{
					Flags:   channel.MessageFlagsEphemeral,
					Content: "Le message sélectionné n'est pas un message de rôles de réaction.",
				},
			}).Do(ctx)
			if err != nil {
				logger.Error("unable to send rolereact message not found", "error", err)
			}
			return
		}
		messageEdits[messageCounter] = target
		targetEditID = messageCounter
		messageCounter++
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseChannelMessageWithSource,
		Data: MessageModifyData(ctx, i.Interaction, &EditID{MessageEditID: targetEditID}),
	}).Do(ctx)
	if err != nil {
		logger.Error("unable to send modify rolereact message", "error", err)
	}
}

func HandleModifyComponent(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditID) {
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: MessageModifyData(ctx, i.Interaction, params),
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send modify rolereact message", "error", err)
	}
}

func HandleResetMessage(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditID) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		cfg := GetGuildConfigPreloaded(ctx, i.GuildID)
		for _, m := range cfg.RrMessages {
			if m.ID == message.ID {
				messageEdits[params.MessageEditID] = &m
			}
		}
		responseData = *MessageModifyData(ctx, i.Interaction, params)
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send reset message message", "error", err)
	}
}

func HandleStartSetNote(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditID) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	if !ok {
		err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send message edit not found message", "error", err)
		}
		return
	}
	resp := interaction.NewModalResponse().
		Title("Changer la description").
		CustomID(dynamicid.FormatCustomID(SetNote, *params)).
		AddComponent(&component.Label{
			Label:       "Nouvelle description",
			Description: "Description affichée sur votre message de réaction",
			Component: &component.TextInput{
				Style:     component.TextInputParagraph,
				MaxLength: 2000,
				CustomID:  "note",
				Value:     message.Note,
			},
		}).
		Response()
	err := dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send edit note modal", "error", err)
	}
}

func HandleSetNote(ctx context.Context, dg bot.Session, i *interaction.ModalSubmit, params *EditID) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	if !ok {
		err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send set note error message", "error", err)
		}
		return
	}
	message.Note = i.Data.Components[0].(*component.Label).Component.(*component.TextInput).Value
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: MessageModifyData(ctx, i.Interaction, params),
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send updated note message", "error", err)
	}
}

func HandleApplyMessage(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditID) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
		err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &responseData,
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send apply message error message", "error", err)
		}
		return
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseDeferredChannelMessageWithSource,
		Data: &interaction.ResponseData{Flags: channel.MessageFlagsEphemeral},
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to defer interaction", "error", err)
		return
	}
	m := ApplyMessageChange(ctx, dg, i.Interaction, message)
	_, err = dg.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
		Content: &m,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send apply rolereaction message changes", "error", err)
	}
}

func HandleNewRole(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditID) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
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
		responseData = MessageModifyRoleData(ctx, i.Interaction, &EditIDWithRole{MessageEditID: params.MessageEditID, RoleCounterID: roleCounter}, "")
		roleCounter++
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send modify reaction role message", "error", err)
	}
}

func HandleOpenRole(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditIDWithRole) {
	_, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	var responseData interaction.ResponseData
	if !ok {
		responseData = interaction.ResponseData{
			Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
			Components: []component.Component{
				&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
			},
		}
	} else {
		responseData = MessageModifyRoleData(ctx, i.Interaction, params, "")
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send open reaction role message", "error", err)
	}
}

func HandleSetRole(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditIDWithRole) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	var responseData interaction.ResponseData
	var role *config.RoleReact
	if ok {
		roleIndex := slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == params.RoleCounterID })
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
		role.RoleID = i.Data.Values[0]
		responseData = MessageModifyRoleData(ctx, i.Interaction, params, "")
	}
	err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send open reaction role message", "error", err)
	}
}

func HandleSetReaction(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditIDWithRole) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	var role *config.RoleReact
	if ok {
		roleIndex := slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == params.RoleCounterID })
		if roleIndex != -1 {
			role = message.Roles[roleIndex]
		}
	}
	if !ok || role == nil {
		err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send open reaction role message", "error", err)
		}
		return
	}
	responseData := MessageModifyRoleData(ctx, i.Interaction, params, "Ajoute la réaction que tu veux choisir au message de rôle de réaction (tu peux y accéder avec le bouton ci-dessous)")
	//TODO: handle errors here
	dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
		Type: types.InteractionResponseUpdateMessage,
		Data: &responseData,
	}).Do(ctx)
	emojiName, ok := WaitForEmoji(ctx, dg, i.Member.User.ID, message.MessageID)
	if !ok {
		editResponseComponents := MessageModifyRoleComponents(ctx, i.Interaction, params, "Le temps d'attente a été dépassé")
		_, err := dg.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Components: &editResponseComponents,
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send timed out reaction message", "error", err)
		}
		return
	}

	err := dg.ChannelAPI().MessageReactionAdd(message.ChannelID, message.MessageID, emojiName).Do(ctx)
	if err != nil {
		editResponseComponents := MessageModifyRoleComponents(ctx, i.Interaction, params, "La réaction n'est pas utilisable. Cela peut être résolu en l'ajoutant à ce serveur")
		_, err := dg.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
			Components: &editResponseComponents,
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send unusable reaction message", "error", err)
		}
		return
	}
	err = dg.ChannelAPI().MessageReactionRemove(message.ChannelID, message.MessageID, emojiName, i.Member.User.ID).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Warn("unable to remove author reaction from message", "error", err)
	}
	role.Reaction = emojiName
	components := MessageModifyRoleComponents(ctx, i.Interaction, params, "")
	_, err = dg.InteractionAPI().ResponseEdit(i.Interaction, &channel.WebhookEdit{
		Flags:      channel.MessageFlagsIsComponentsV2 | channel.MessageFlagsEphemeral,
		Components: &components,
	}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to edit original response", "error", err)
	}
}

func HandleDelRole(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *EditIDWithRole) {
	message, ok := GetMessageFromEditID(ctx, i.Interaction, params.MessageEditID)
	roleIndex := -1
	if ok {
		roleIndex = slices.IndexFunc(message.Roles, func(role *config.RoleReact) bool { return role.CounterID == params.RoleCounterID })
	}
	if !ok || roleIndex == -1 {
		err := dg.InteractionAPI().Respond(i.Interaction, &interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: &interaction.ResponseData{
				Flags: channel.MessageFlagsEphemeral | channel.MessageFlagsIsComponentsV2,
				Components: []component.Component{
					&component.TextDisplay{Content: "Impossible de trouver la modification de message. Veuillez réessayer."},
				},
			},
		}).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("unable to send open reaction role message", "error", err)
		}
		return
	}
	message.Roles = append(message.Roles[:roleIndex],
		message.Roles[roleIndex+1:]...,
	)
	err := dg.InteractionAPI().Respond(i.Interaction,
		&interaction.Response{
			Type: types.InteractionResponseUpdateMessage,
			Data: MessageModifyData(ctx, i.Interaction, &EditID{MessageEditID: params.MessageEditID}),
		}).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("unable to send modify message message", "error", err)
	}
}
