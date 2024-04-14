package xp

import (
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
)

func OnMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	c := Copaing{DiscordID: m.Author.ID, GuildID: m.GuildID}
	c.Load()
	// add xp
	pastLevel := Level(c.XP)
	trimmed := utils.TrimMessage(m.Content)
	c.XP += XPMessage(uint(len(trimmed)), calcDiversity(trimmed))
	c.Save()
	newLevel := Level(c.XP)
	// handle new level
	if pastLevel < newLevel {
		if err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, "⬆"); err != nil {
			utils.SendAlert("xp/events.go - reaction add new level", "cannot add the reaction: "+err.Error())
		}
		onNewLevel(s, newLevel)
	}
}

func calcDiversity(msg string) uint {
	var chars []rune
	for _, c := range []rune(msg) {
		toAdd := true
		for _, ch := range chars {
			if ch == c {
				toAdd = false
			}
		}
		if toAdd {
			chars = append(chars, c)
		}
	}
	return uint(len(chars))
}
