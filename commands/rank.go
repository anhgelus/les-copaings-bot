package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
)

func Rank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optMap := utils.GenerateOptionMap(i)
	c := xp.GetCopaing(i.Member.User.ID, i.GuildID) // current copaing = member who used /rank
	xp.LastEventUpdate(s, c)                        // update xp and reset last event
	msg := "Votre niveau"
	m := i.Member
	var err error
	resp := utils.ResponseBuilder{C: s, I: i}
	if v, ok := optMap["copaing"]; ok {
		u := v.UserValue(s)
		if u.Bot {
			err = resp.Message("Imagine si les bots avaient un niveau :rolling_eyes:").IsEphemeral().Send()
			if err != nil {
				utils.SendAlert("rank.go - Reply error user is a bot", err.Error())
			}
		}
		m, err = s.GuildMember(i.GuildID, u.ID)
		if err != nil {
			utils.SendAlert(
				"rank.go - Fetching guild member",
				err.Error(),
				"discord_id",
				u.ID,
				"guild_id",
				i.GuildID,
			)
			err = resp.Message("Erreur : impossible de récupérer le membre").IsEphemeral().Send()
			if err != nil {
				utils.SendAlert("rank.go - Reply error fetching guild member", err.Error())
			}
			return
		}
		c.DiscordID = u.ID // current copaing = member targeted by member who wrote /rank
		err = c.Load()     // reload copaing (change line before)
		if err != nil {
			utils.SendAlert(
				"commands/rank.go - Loading copaing",
				err.Error(),
				"discord_id",
				u.ID,
				"guild_id",
				i.GuildID,
			)
			return
		}
		xp.XPUpdate(s, c) // update xp without resetting event
		msg = fmt.Sprintf("Le niveau de %s", m.DisplayName())
	}
	lvl := xp.Level(c.XP)
	nxtLvlXP := xp.XPForLevel(lvl + 1)
	err = resp.Message(fmt.Sprintf(
		"%s : **%d**\n> XP : %d\n> Prochain niveau dans %d XP",
		msg,
		lvl,
		c.XP,
		nxtLvlXP-c.XP,
	)).Send()
	if err != nil {
		utils.SendAlert("rank.go - Sending rank", err.Error())
	}
}
