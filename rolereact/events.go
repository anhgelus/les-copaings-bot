package rolereact

import (
	"context"

	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/guild"
)

type RoleReact struct {
	RoleID string
}

func HandleReactionAdd(ctx context.Context, dg bot.Session, e *event.MessageReactionAdd) {
	results := []RoleReact{}
	common.GetDB(ctx).Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	for _, role := range results {
		err := guild.AddMemberRole(e.GuildID, e.UserID, role.RoleID).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("Unable to add role after member added reaction", "error", err)
		}
	}
}

func HandleReactionRemove(ctx context.Context, dg bot.Session, e *event.MessageReactionRemove) {
	results := []RoleReact{}
	common.GetDB(ctx).Model(&config.RoleReact{}).
		Joins("JOIN role_react_messages ON role_reacts.role_react_message_id = role_react_messages.id").
		Where("role_react_messages.message_id = ? AND role_reacts.reaction = ?", e.MessageID, e.MessageReaction.Emoji.APIName()).
		Scan(&results)
	for _, role := range results {
		err := guild.RemoveMemberRole(e.GuildID, e.UserID, role.RoleID).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("Unable to remove role after member removed reaction", "error", err)
		}
	}
}
