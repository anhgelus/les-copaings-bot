package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	xp2 "github.com/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	ConnectedSince  = "connected_since"
	NotConnected    = -1
	MaxTimeInVocal  = 60 * 60 * 6
	MaxXpPerMessage = 250
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
	user.LastEventUpdate(s, c)
	// add exp
	trimmed := utils.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	xp := xp2.MessageXP(uint(len(trimmed)), calcDiversity(trimmed))
	if xp > MaxXpPerMessage {
		xp = MaxXpPerMessage
	}
	c.AddXP(s, m.Member, xp, func(_ uint, _ uint) {
		if err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			utils.SendAlert(
				"exp/events.go - add reaction for new level", err.Error(),
				"channel id", m.ChannelID,
				"message id", m.Message.ID,
			)
		}
	})
}

func calcDiversity(msg string) uint {
	var chars []rune
	for _, c := range []rune(msg) {
		if !slices.Contains(chars, c) {
			chars = append(chars, c)
		}
	}
	return uint(len(chars))
}

func OnVoiceUpdate(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
	if e.Member.User.Bot {
		return
	}
	user.LastEventUpdate(s, user.GetCopaing(e.UserID, e.GuildID))
	cfg := config.GetGuildConfig(e.GuildID)
	client, err := config.GetRedisClient()
	if err != nil {
		utils.SendAlert("exp/events.go - Getting redis client", err.Error())
		return
	}
	if e.BeforeUpdate == nil && e.ChannelID != "" {
		if cfg.IsDisabled(e.ChannelID) {
			return
		}
		onConnection(s, e, client)
	} else if e.BeforeUpdate != nil && e.ChannelID == "" {
		if cfg.IsDisabled(e.BeforeUpdate.ChannelID) {
			return
		}
		onDisconnect(s, e, client)
	}
}

func onConnection(_ *discordgo.Session, e *discordgo.VoiceStateUpdate, client *redis.Client) {
	utils.SendDebug("User connected", "username", e.Member.DisplayName())
	c := user.GetCopaing(e.UserID, e.GuildID)
	err := client.Set(
		context.Background(),
		c.GenKey(ConnectedSince),
		strconv.FormatInt(time.Now().Unix(), 10),
		0,
	).Err()
	if err != nil {
		utils.SendAlert("exp/events.go - Setting connected_since", err.Error())
	}
}

func onDisconnect(s *discordgo.Session, e *discordgo.VoiceStateUpdate, client *redis.Client) {
	now := time.Now().Unix()
	c := user.GetCopaing(e.UserID, e.GuildID)
	key := c.GenKey(ConnectedSince)
	res := client.Get(context.Background(), key)
	// check validity of user (1)
	if errors.Is(res.Err(), redis.Nil) {
		utils.SendWarn(fmt.Sprintf(
			"User %s diconnect from a vocal but does not have a connected_since", e.Member.DisplayName(),
		))
		return
	}
	if res.Err() != nil {
		utils.SendAlert("exp/events.go - Getting connected_since", res.Err().Error())
		err := client.Set(context.Background(), key, strconv.Itoa(NotConnected), 0).Err()
		if err != nil {
			utils.SendAlert("exp/events.go - Set connected_since to not connected after get err", err.Error())
		}
		return
	}
	con, err := res.Int64()
	if err != nil {
		utils.SendAlert("exp/events.go - Converting result to int64", err.Error())
		return
	}
	// check validity of user (2)
	if con == NotConnected {
		utils.SendWarn(fmt.Sprintf(
			"User %s diconnect from a vocal but was registered as not connected", e.Member.DisplayName(),
		))
		return
	}
	utils.SendDebug("User disconnected", "username", e.Member.DisplayName(), "since", con)
	err = client.Set(context.Background(), key, strconv.Itoa(NotConnected), 0).Err()
	if err != nil {
		utils.SendAlert("exp/events.go - Set connected_since to not connected", err.Error())
	}
	// add exp
	timeInVocal := now - con
	if timeInVocal < 0 {
		utils.SendAlert("exp/events.go - Calculating time spent in vocal", "the time is negative")
		return
	}
	if timeInVocal > MaxTimeInVocal {
		utils.SendWarn(fmt.Sprintf("User %s spent more than 6 hours in vocal", e.Member.DisplayName()))
		timeInVocal = MaxTimeInVocal
	}
	e.Member.GuildID = e.GuildID
	c.AddXP(s, e.Member, xp2.VocalXP(uint(timeInVocal)), func(_ uint, newLevel uint) {
		cfg := config.GetGuildConfig(e.GuildID)
		_, err = s.ChannelMessageSend(cfg.FallbackChannel, fmt.Sprintf(
			"%s est maintenant niveau %d", e.Member.Mention(), newLevel,
		))
		if err != nil {
			utils.SendAlert("exp/events.go - Sending new level in fallback channel", err.Error())
		}
	})
}

func OnLeave(_ *discordgo.Session, e *discordgo.GuildMemberRemove) {
	utils.SendDebug("Leave event", "user_id", e.User.ID)
	c := user.GetCopaing(e.User.ID, e.GuildID)
	if err := gokord.DB.Where("guild_id = ?", e.GuildID).Delete(c).Error; err != nil {
		utils.SendAlert(
			"exp/events.go - deleting user from db", err.Error(),
			"user_id", e.User.ID,
			"guild_id", e.GuildID,
		)
	}
}
