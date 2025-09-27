package rolereact

import (
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	oldGokord "github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"
)

type RoleReact struct {
	RoleID string
}

func HandleReactionAdd(
	s bot.Session,
	e *event.MessageReactionAdd,
) {
	results := []RoleReact{}
	oldGokord.DB.Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	for _, role := range results {
		err := s.GuildAPI().MemberRoleAdd(e.GuildID, e.UserID, role.RoleID)
		if err != nil {
			s.Logger().Error("Unable to add role after member added reaction", "error", err)
		}
	}
}

func HandleReactionRemove(
	s bot.Session,
	e *event.MessageReactionRemove,
) {
	results := []RoleReact{}
	oldGokord.DB.Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	for _, role := range results {
		err := s.GuildAPI().MemberRoleRemove(e.GuildID, e.UserID, role.RoleID)
		if err != nil {
			s.Logger().Error("Unable to remove role after member removed reaction", "error", err)
		}
	}
}
