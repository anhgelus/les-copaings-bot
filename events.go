package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/channel"
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

func OnMessage(ctx context.Context, dg bot.Session, m *event.MessageCreate) {
	if m.Author.Bot {
		return
	}
	cfg := config.GetGuildConfig(ctx, m.GuildID)
	if cfg.IsDisabled(ctx, dg, m.ChannelID) {
		return
	}
	cc := user.GetCopaing(ctx, m.Author.ID, m.GuildID)
	// add exp
	trimmed := exp.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	xp := min(exp.MessageXP(uint(len(trimmed)), exp.CalcDiversity(trimmed)), MaxXpPerMessage)
	cc.AddXP(ctx, dg, m.Member, xp, func(_ uint, _ uint) {
		if err := channel.AddReaction(m.ChannelID, m.Message.ID, "⬆").Do(ctx); err != nil {
			bot.Logger(ctx).Error(
				"add reaction for new level",
				"error", err, "channel", m.ChannelID, "message", m.Message.ID,
			)
		}
	})
}

func OnVoiceUpdate(ctx context.Context, dg bot.Session, e *event.VoiceStateUpdate) {
	if e.Member.User.Bot {
		return
	}
	cfg := config.GetGuildConfig(ctx, e.GuildID)
	dis := cfg.IsDisabled(ctx, dg, e.BeforeUpdate.ChannelID)
	if (e.BeforeUpdate == nil || dis) && e.ChannelID != "" {
		if dis {
			return
		}
		onConnection(ctx, dg, e)
	} else if e.BeforeUpdate != nil && (e.ChannelID == "" || dis) {
		if dis {
			return
		}
		onDisconnect(ctx, dg, e)
	}
}

func genMapKey(guildID string, userID string) string {
	return fmt.Sprintf("%s:%s", guildID, userID)
}

func onConnection(ctx context.Context, _ bot.Session, e *event.VoiceStateUpdate) {
	bot.Logger(ctx).Debug("user connected", "user", e.Member.DisplayName())
	connectedSince[genMapKey(e.GuildID, e.UserID)] = time.Now().Unix()
}

func onDisconnect(ctx context.Context, dg bot.Session, e *event.VoiceStateUpdate) {
	now := time.Now().Unix()
	cc := user.GetCopaing(ctx, e.UserID, e.GuildID)
	// check the validity of user
	con, ok := connectedSince[genMapKey(e.GuildID, e.UserID)]
	if !ok || con == NotConnected {
		bot.Logger(ctx).Warn(
			"user disconnect from a vocal but was registered as not connected",
			"user", e.Member.DisplayName(),
		)
		return
	}
	timeInVocal := now - con
	bot.Logger(ctx).Debug("user disconnected", "user", e.Member.DisplayName(), "time in vocal", timeInVocal)
	connectedSince[genMapKey(e.GuildID, e.UserID)] = NotConnected
	// add exp
	if timeInVocal < 0 {
		bot.Logger(ctx).Warn("time spent in vocal is negative", "user", e.Member.DisplayName(), "guild", e.GuildID)
		return
	}
	timeInVocal = min(timeInVocal, MaxTimeInVocal)
	e.Member.GuildID = e.GuildID
	cc.AddXP(ctx, dg, e.Member, exp.VocalXP(uint(timeInVocal)), func(_ uint, newLevel uint) {
		cfg := config.GetGuildConfig(ctx, e.GuildID)
		if len(cfg.FallbackChannel) == 0 {
			return
		}
		_, err := channel.SendMessage(cfg.FallbackChannel, fmt.Sprintf(
			"%s est maintenant niveau %d", e.Member.Mention(), newLevel,
		)).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("sending new level in fallback channel", "error", err)
		}
	})
}

func OnLeave(ctx context.Context, dg bot.Session, e *event.GuildMemberRemove) {
	bot.Logger(ctx).Debug("leave event", "user", e.User.Username)
	if e.User.Bot {
		return
	}
	c := user.GetCopaing(ctx, e.User.ID, e.GuildID)
	if err := c.Delete(ctx); err != nil {
		bot.Logger(ctx).Error("deleting user", "user", e.User.Username, "guild", e.GuildID)
	}
}
