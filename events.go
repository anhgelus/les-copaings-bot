package main

import (
	"fmt"
	"strings"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord/logger"
	"github.com/bwmarrin/discordgo"
)

const (
	NotConnected    = -1
	MaxTimeInVocal  = 60 * 60 * 6
	MaxXpPerMessage = 250
)

var (
	connectedSince = map[string]int64{}
)

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	cfg := config.GetGuildConfig(m.GuildID)
	if cfg.IsDisabled(m.ChannelID) {
		return
	}
	c := user.GetCopaing(m.Author.ID, m.GuildID)
	// add exp
	trimmed := exp.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	xp := min(exp.MessageXP(uint(len(trimmed)), exp.CalcDiversity(trimmed)), MaxXpPerMessage)
	c.AddXP(s, m.Member, xp, func(_ uint, _ uint) {
		if err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			logger.Alert(
				"events.go - add reaction for new level", err.Error(),
				"channel id", m.ChannelID,
				"message id", m.Message.ID,
			)
		}
	})
}

func OnVoiceUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.Member.User.Bot {
		return
	}
	cfg := config.GetGuildConfig(e.GuildID)
	if (e.BeforeUpdate == nil || cfg.IsDisabled(e.BeforeUpdate.ChannelID)) && e.ChannelID != "" {
		if cfg.IsDisabled(e.ChannelID) {
			return
		}
		onConnection(s, e)
	} else if e.BeforeUpdate != nil && (e.ChannelID == "" || cfg.IsDisabled(e.ChannelID)) {
		if cfg.IsDisabled(e.BeforeUpdate.ChannelID) {
			return
		}
		onDisconnect(s, e)
	}
}

func genMapKey(guildID string, userID string) string {
	return fmt.Sprintf("%s:%s", guildID, userID)
}

func onConnection(_ *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	logger.Debug("User connected", "username", e.Member.DisplayName())
	connectedSince[genMapKey(e.GuildID, e.UserID)] = time.Now().Unix()
}

func onDisconnect(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	now := time.Now().Unix()
	c := user.GetCopaing(e.UserID, e.GuildID)
	// check the validity of user
	con, ok := connectedSince[genMapKey(e.GuildID, e.UserID)]
	if !ok || con == NotConnected {
		logger.Warn(fmt.Sprintf(
			"User %s diconnect from a vocal but was registered as not connected", e.Member.DisplayName(),
		))
		return
	}
	timeInVocal := now - con
	logger.Debug("User disconnected", "username", e.Member.DisplayName(), "time in vocal", timeInVocal)
	connectedSince[genMapKey(e.GuildID, e.UserID)] = NotConnected
	// add exp
	if timeInVocal < 0 {
		logger.Alert(
			"events.go - Calculating time spent in vocal", "the time is negative",
			"discord_id", e.UserID,
			"guild_id", e.GuildID,
		)
		return
	}
	if timeInVocal > MaxTimeInVocal {
		logger.Warn(fmt.Sprintf("User %s spent more than 6 hours in vocal", e.Member.DisplayName()))
		timeInVocal = MaxTimeInVocal
	}
	e.Member.GuildID = e.GuildID
	c.AddXP(s, e.Member, exp.VocalXP(uint(timeInVocal)), func(_ uint, newLevel uint) {
		cfg := config.GetGuildConfig(e.GuildID)
		if len(cfg.FallbackChannel) == 0 {
			return
		}
		_, err := s.ChannelMessageSend(cfg.FallbackChannel, fmt.Sprintf(
			"%s est maintenant niveau %d", e.Member.Mention(), newLevel,
		))
		if err != nil {
			logger.Alert("events.go - Sending new level in fallback channel", err.Error())
		}
	})
}

func OnLeave(_ *discordgo.Session, e *discordgo.GuildMemberRemove) {
	logger.Debug("Leave event", "user_id", e.User.ID)
	if e.User.Bot {
		return
	}
	c := user.GetCopaing(e.User.ID, e.GuildID)
	if err := c.Delete(); err != nil {
		logger.Alert(
			"events.go - deleting user from db", err.Error(),
			"user_id", e.User.ID,
			"guild_id", e.GuildID,
		)
	}
}
