package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/event"
)

const (
	NotConnected    = -1
	MaxTimeInVocal  = 60 * 60 * 6
	MaxXpPerMessage = 250
)

var (
	connectedSince = map[string]int64{}
)

func OnMessage(ctx context.Context, s bot.Session, m *event.MessageCreate) {
	if m.Author.Bot {
		return
	}
	cfg := config.GetGuildConfig(m.GuildID)
	if cfg.IsDisabled(s, m.ChannelID) {
		return
	}
	cc := user.GetCopaing(ctx, m.Author.ID, m.GuildID)
	// add exp
	trimmed := exp.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	xp := min(exp.MessageXP(uint(len(trimmed)), exp.CalcDiversity(trimmed)), MaxXpPerMessage)
	cc.AddXP(ctx, s, m.Member, xp, func(_ uint, _ uint) {
		if err := s.ChannelAPI().MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			s.Logger().Error(
				"add reaction for new level",
				"error", err,
				"channel", m.ChannelID,
				"message", m.Message.ID,
			)
		}
	})
}

func OnVoiceUpdate(ctx context.Context, s bot.Session, e *event.VoiceStateUpdate) {
	if e.Member.User.Bot {
		return
	}
	cfg := config.GetGuildConfig(e.GuildID)
	if (e.BeforeUpdate == nil || cfg.IsDisabled(s, e.BeforeUpdate.ChannelID)) && e.ChannelID != "" {
		if cfg.IsDisabled(s, e.ChannelID) {
			return
		}
		onConnection(s, e)
	} else if e.BeforeUpdate != nil && (e.ChannelID == "" || cfg.IsDisabled(s, e.ChannelID)) {
		if cfg.IsDisabled(s, e.BeforeUpdate.ChannelID) {
			return
		}
		onDisconnect(ctx, s, e)
	}
}

func genMapKey(guildID string, userID string) string {
	return fmt.Sprintf("%s:%s", guildID, userID)
}

func onConnection(s bot.Session, e *event.VoiceStateUpdate) {
	s.Logger().Debug("user connected", "user", e.Member.DisplayName())
	connectedSince[genMapKey(e.GuildID, e.UserID)] = time.Now().Unix()
}

func onDisconnect(ctx context.Context, s bot.Session, e *event.VoiceStateUpdate) {
	now := time.Now().Unix()
	cc := user.GetCopaing(ctx, e.UserID, e.GuildID)
	// check the validity of user
	con, ok := connectedSince[genMapKey(e.GuildID, e.UserID)]
	if !ok || con == NotConnected {
		s.Logger().Warn(
			"user disconnect from a vocal but was registered as not connected",
			"user", e.Member.DisplayName(),
		)
		return
	}
	timeInVocal := now - con
	s.Logger().Debug("user disconnected", "user", e.Member.DisplayName(), "time in vocal", timeInVocal)
	connectedSince[genMapKey(e.GuildID, e.UserID)] = NotConnected
	// add exp
	if timeInVocal < 0 {
		s.Logger().Warn("time spent in vocal is negative", "user", e.Member.DisplayName(), "guild", e.GuildID)
		return
	}
	timeInVocal = min(timeInVocal, MaxTimeInVocal)
	e.Member.GuildID = e.GuildID
	cc.AddXP(ctx, s, e.Member, exp.VocalXP(uint(timeInVocal)), func(_ uint, newLevel uint) {
		cfg := config.GetGuildConfig(e.GuildID)
		if len(cfg.FallbackChannel) == 0 {
			return
		}
		_, err := s.ChannelAPI().MessageSend(cfg.FallbackChannel, fmt.Sprintf(
			"%s est maintenant niveau %d", e.Member.Mention(), newLevel,
		))
		if err != nil {
			s.Logger().Error("sending new level in fallback channel", "error", err)
		}
	})
}

func OnLeave(ctx context.Context, s bot.Session, e *event.GuildMemberRemove) {
	s.Logger().Debug("leave event", "user", e.User.Username)
	if e.User.Bot {
		return
	}
	c := user.GetCopaing(ctx, e.User.ID, e.GuildID).Copaing(ctx)
	err := gokord.DB.
		Where("copaing_id = ? and guild_id = ?", c.ID, e.GuildID).
		Delete(&user.CopaingXP{}).
		Error
	if err != nil {
		s.Logger().Error("deleting user xp from DB", "user", e.User.Username, "guild", e.GuildID)
	}
	if err = c.Delete(ctx); err != nil {
		s.Logger().Error("deleting user from DB", "user", e.User.Username, "guild", e.GuildID)
	}
}
