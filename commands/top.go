package commands

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
)

func Top(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user.LastEventUpdate(s, user.GetCopaing(i.Member.User.ID, i.GuildID))
	resp := utils.ResponseBuilder{C: s, I: i}
	err := resp.IsDeferred().Send()
	if err != nil {
		utils.SendAlert("commands/top.go - Sending deferred", err.Error())
		return
	}
	resp.NotDeferred().IsEdit()
	go func() {
		var tops []user.Copaing
		gokord.DB.Where("guild_id = ?", i.GuildID).Limit(10).Order("exp desc").Find(&tops)
		msg := ""
		for i, c := range tops {
			if i == 9 {
				msg += fmt.Sprintf("%d. **<@%s>** - niveau %d", i+1, c.DiscordID, exp.Level(c.XP))
			} else {
				msg += fmt.Sprintf("%d. **<@%s>** - niveau %d\n", i+1, c.DiscordID, exp.Level(c.XP))
			}
		}
		err = resp.Embeds([]*discordgo.MessageEmbed{
			{
				Title:       "Top",
				Description: msg,
				Color:       utils.Success,
			},
		}).Send()
		if err != nil {
			utils.SendAlert("commands/top.go - Sending response top", err.Error())
		}
	}()
}
