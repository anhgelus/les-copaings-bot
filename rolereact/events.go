package rolereact

import (
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	oldGokord "github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord"
)

type RoleReact struct {
	RoleID string
}

func HandleReactionAdd(
	s *gokord.Session,
	e *gokord.MessageReactionAdd,
) {
	results := []RoleReact{}
	oldGokord.DB.Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	s.LogDebug("test: %#v\n", results)
	for _, role := range results {
		err := s.GuildAPI().MemberRoleAdd(e.GuildID, e.UserID, role.RoleID)
		if err != nil {
			s.LogError(err, "Unable to add message after member added reaction")
		}
	}
}

func HandleReactionRemove(
	s *gokord.Session,
	e *gokord.MessageReactionRemove,
) {
	results := []RoleReact{}
	oldGokord.DB.Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	s.LogDebug("test: %#v\n", results)
	for _, role := range results {
		s.GuildAPI().MemberRoleRemove(e.GuildID, e.UserID, role.RoleID)
	}
}
