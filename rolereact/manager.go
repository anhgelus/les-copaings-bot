package rolereact

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	oldGokord "github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
	"github.com/nyttikord/gokord/emoji"
	"github.com/nyttikord/gokord/event"
)

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

func ApplyMessageChange(s bot.Session, i *event.InteractionCreate, message *config.RoleReactMessage) string {
	messageContent := MessageContent(message)
	_, err := s.ChannelAPI().MessageEditComplex(
		&channel.MessageEdit{
			Content:         &messageContent,
			AllowedMentions: &channel.MessageAllowedMentions{},
			Channel:         message.ChannelID,
			ID:              message.MessageID,
		},
	)
	if err != nil {
		s.Logger().Error("unable to update rolereact message", "error", err)
		return "Impossible de mettre à jour le message."
	}
	for _, role := range message.Roles {
		if role.Reaction != "" && role.RoleID != "" && err == nil {
			err = s.ChannelAPI().MessageReactionAdd(
				message.ChannelID,
				message.MessageID,
				role.Reaction,
			)
		}
	}
	if err != nil {
		s.Logger().Error("unable to update reactions on rolereact message", "error", err)
		return "Impossible de mettre à jour le message."
	}
	cfg := GetGuildConfigPreloaded(i.GuildID)
	messageIndex := slices.IndexFunc(cfg.RrMessages, func(m config.RoleReactMessage) bool { return m.ID == message.ID })
	if messageIndex != -1 {
		oldMessage := cfg.RrMessages[messageIndex]
		roles := make(map[uint]config.RoleReact, len(message.Roles))
		for _, role := range message.Roles {
			roles[role.ID] = *role
		}
		for _, role := range oldMessage.Roles {
			_, ok := roles[role.ID]
			if !ok {
				err := oldGokord.DB.Delete(role).Error
				if err != nil {
					s.Logger().Error("unable to delete reaction role from database", "error", err)
					return "Impossible de sauvegarder le message de rôle. Merci de contacter l'administrateur du bot."
				}
			}
		}
		cfg.RrMessages[messageIndex] = *message
		err := oldGokord.DB.Save(cfg.RrMessages[messageIndex]).Error
		if err != nil {
			s.Logger().Error("unable to save rolereaction message in database", "error", err)
			return "Impossible de sauvegarder le message de rôle. Merci de contacter l'administrateur du bot."
		}
		for _, role := range cfg.RrMessages[messageIndex].Roles {
			err = oldGokord.DB.Save(role).Error
			if err != nil {
				s.Logger().Error("unable to save rolereaction role in database", "error", err)
				return "Impossible de sauvegarder le message de rôle. Merci de contacter l'administrateur du bot."
			}
		}
	}
	return "Message de réaction mis à jour avec succès !"
}

func WaitForEmoji(s bot.Session, userID string, messageID string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	emojiChann := make(chan emoji.Emoji)

	cancelHandler := s.EventManager().AddHandler(func(s bot.Session, e *event.MessageReactionAdd) {
		if e.MessageID == messageID && e.UserID == userID {
			emojiChann <- e.Emoji
		}
	})
	defer cancelHandler()

	select {
	case emoji := <-emojiChann:
		emojiName := emoji.APIName()
		return emojiName, true
	case <-ctx.Done():
		return "", false
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
