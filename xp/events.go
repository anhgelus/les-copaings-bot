package xp

import (
	"context"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	ConnectedSince = "connected_since"
	NotConnected   = -1
	MaxTimeInVocal = 60 * 60 * 6
)

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	cfg := config.GetGuildConfig(m.GuildID)
	if cfg.IsDisabled(m.ChannelID) {
		return
	}
	c := GetCopaing(m.Author.ID, m.GuildID)
	// add xp
	trimmed := utils.TrimMessage(strings.ToLower(m.Content))
	m.Member.User = m.Author
	m.Member.GuildID = m.GuildID
	c.AddXP(s, m.Member, XPMessage(uint(len(trimmed)), calcDiversity(trimmed)), func(_ uint, _ uint) {
		if err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			utils.SendAlert(
				"xp/events.go - add reaction for new level",
				err.Error(),
				"channel id",
				m.ChannelID,
				"message id",
				m.Message.ID,
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
	client, err := getRedisClient()
	if err != nil {
		utils.SendAlert("xp/events.go - Getting redis client", err.Error())
		return
	}
	if e.BeforeUpdate == nil {
		onConnection(s, e, client)
	} else {
		onDisconnect(s, e, client)
	}
}

func onConnection(_ *discordgo.Session, e *discordgo.VoiceStateUpdate, client *redis.Client) {
	u := gokord.UserBase{DiscordID: e.UserID, GuildID: e.GuildID}
	err := client.Set(
		context.Background(),
		fmt.Sprintf("%s:%s", u.GenKey(), ConnectedSince),
		strconv.FormatInt(time.Now().Unix(), 10),
		0,
	).Err()
	if err != nil {
		utils.SendAlert("xp/events.go - Setting connected_since", err.Error())
	}
}

func onDisconnect(s *discordgo.Session, e *discordgo.VoiceStateUpdate, client *redis.Client) {
	now := time.Now().Unix()
	u := gokord.UserBase{DiscordID: e.UserID, GuildID: e.GuildID}
	key := fmt.Sprintf("%s:%s", u.GenKey(), ConnectedSince)
	res := client.Get(context.Background(), key)
	// check validity of user (1)
	if errors.Is(res.Err(), redis.Nil) {
		utils.SendWarn(fmt.Sprintf(
			"User %s diconnect from a vocal but does not have a connected_since",
			e.Member.DisplayName(),
		))
		return
	}
	if res.Err() != nil {
		utils.SendAlert("xp/events.go - Getting connected_since", res.Err().Error())
		err := client.Set(context.Background(), key, strconv.Itoa(NotConnected), 0).Err()
		if err != nil {
			utils.SendAlert("xp/events.go - Set connected_since to not connected after get err", err.Error())
		}
		return
	}
	con, err := res.Int64()
	if err != nil {
		utils.SendAlert("xp/events.go - Converting result to int64", err.Error())
		return
	}
	// check validity of user (2)
	if con == NotConnected {
		utils.SendWarn(fmt.Sprintf(
			"User %s diconnect from a vocal but was registered as not connected",
			e.Member.DisplayName(),
		))
		return
	}
	err = client.Set(context.Background(), key, strconv.Itoa(NotConnected), 0).Err()
	if err != nil {
		utils.SendAlert("xp/events.go - Set connected_since to not connected", err.Error())
	}
	// add xp
	timeInVocal := now - con
	if timeInVocal < 0 {
		utils.SendAlert("xp/events.go - Calculating time spent in vocal", "the time is negative")
		return
	}
	if timeInVocal > MaxTimeInVocal {
		utils.SendWarn(fmt.Sprintf("User %s spent more than 6 hours in vocal", e.Member.DisplayName()))
		timeInVocal = MaxTimeInVocal
	}
	c := GetCopaing(u.DiscordID, u.GuildID)
	e.Member.GuildID = e.GuildID
	c.AddXP(s, e.Member, XPVocal(uint(timeInVocal)), func(_ uint, _ uint) {
		//TODO: handle new level in vocal
	})
}
