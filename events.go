package main

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
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
	trimmed := utils.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	xp := min(exp.MessageXP(uint(len(trimmed)), exp.CalcDiversity(trimmed)), MaxXpPerMessage)
	c.AddXP(s, m.Member, xp, func(_ uint, _ uint) {
		if err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			utils.SendAlert(
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
	if e.BeforeUpdate == nil && e.ChannelID != "" {
		if cfg.IsDisabled(e.ChannelID) {
			return
		}
		onConnection(s, e)
	} else if e.BeforeUpdate != nil && e.ChannelID == "" {
		if cfg.IsDisabled(e.BeforeUpdate.ChannelID) {
			return
		}
		onDisconnect(s, e)
	}
}

func onConnection(_ *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	utils.SendDebug("User connected", "username", e.Member.DisplayName())
	connectedSince[e.UserID] = time.Now().Unix()
}

func onDisconnect(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	now := time.Now().Unix()
	c := user.GetCopaing(e.UserID, e.GuildID)
	// check the validity of user
	con := connectedSince[e.UserID]
	if con == NotConnected {
		utils.SendWarn(fmt.Sprintf(
			"User %s diconnect from a vocal but was registered as not connected", e.Member.DisplayName(),
		))
		return
	}
	timeInVocal := now - con
	utils.SendDebug("User disconnected", "username", e.Member.DisplayName(), "time in vocal", timeInVocal)
	connectedSince[e.UserID] = NotConnected
	// add exp
	if timeInVocal < 0 {
		utils.SendAlert(
			"events.go - Calculating time spent in vocal", "the time is negative",
			"discord_id", e.UserID,
			"guild_id", e.GuildID,
		)
		return
	}
	if timeInVocal > MaxTimeInVocal {
		utils.SendWarn(fmt.Sprintf("User %s spent more than 6 hours in vocal", e.Member.DisplayName()))
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
			utils.SendAlert("events.go - Sending new level in fallback channel", err.Error())
		}
	})
}

func OnLeave(_ *discordgo.Session, e *discordgo.GuildMemberRemove) {
	utils.SendDebug("Leave event", "user_id", e.User.ID)
	if e.User.Bot {
		return
	}
	c := user.GetCopaing(e.User.ID, e.GuildID)
	if err := c.Delete(); err != nil {
		utils.SendAlert(
			"events.go - deleting user from db", err.Error(),
			"user_id", e.User.ID,
			"guild_id", e.GuildID,
		)
	}
}
