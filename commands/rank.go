package commands

import (
	"fmt"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
)

func Rank(s *discordgo.Session, i *discordgo.InteractionCreate) {
	optMap := utils.GenerateOptionMap(i)
	c := user.GetCopaing(i.Member.User.ID, i.GuildID) // current user = member who used /rank
	msg := "Votre niveau"
	m := i.Member
	var err error
	resp := utils.ResponseBuilder{C: s, I: i}
	if v, ok := optMap["user"]; ok {
		u := v.UserValue(s)
		if u.Bot {
			err = resp.Message("Imagine si les bots avaient un niveau :rolling_eyes:").IsEphemeral().Send()
			if err != nil {
				utils.SendAlert("commands/rank.go - Reply error user is a bot", err.Error())
			}
		}
		m, err = s.GuildMember(i.GuildID, u.ID)
		if err != nil {
			utils.SendAlert(
				"commands/rank.go - Fetching guild member",
				err.Error(),
				"discord_id",
				u.ID,
				"guild_id",
				i.GuildID,
			)
			err = resp.Message("Erreur : impossible de récupérer le membre").IsEphemeral().Send()
			if err != nil {
				utils.SendAlert("commands/rank.go - Reply error fetching guild member", err.Error())
			}
			return
		}
		c = user.GetCopaing(u.ID, i.GuildID) // current user = member targeted by member who wrote /rank
		msg = fmt.Sprintf("Le niveau de %s", m.DisplayName())
	}
	xp, err := c.GetXP()
	if err != nil {
		utils.SendAlert(
			"commands/rank.go - Fetching XP",
			err.Error(),
			"discord_id",
			c.ID,
			"guild_id",
			i.GuildID,
		)
		err = resp.Message("Erreur : impossible de récupérer l'XP").IsEphemeral().Send()
		if err != nil {
			utils.SendAlert("commands/rank.go - Reply error fetching xp", err.Error())
		}
		return
	}
	lvl := exp.Level(xp)
	nxtLvlXP := exp.LevelXP(lvl + 1)
	err = resp.Message(fmt.Sprintf(
		"%s : **%d**\n> XP : %d\n> Prochain niveau dans %d XP",
		msg,
		lvl,
		xp,
		nxtLvlXP-xp,
	)).Send()
	if err != nil {
		utils.SendAlert("commands/rank.go - Sending rank", err.Error())
	}
}
