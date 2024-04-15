package commands

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
)

func Top(s *discordgo.Session, i *discordgo.InteractionCreate) {
	xp.LastEventUpdate(xp.GetCopaing(i.User.ID, i.GuildID))
	resp := utils.ResponseBuilder{C: s, I: i}
	err := resp.IsDeferred().Send()
	if err != nil {
		utils.SendAlert("commands/top.go - Sending deferred", err.Error())
		return
	}
	resp.NotDeferred().IsEdit()
	go func() {
		var tops []xp.Copaing
		gokord.DB.Where("guild_id = ?", i.GuildID).Limit(10).Order("xp desc").Find(&tops)
		msg := ""
		for i, c := range tops {
			if i == 9 {
				msg += fmt.Sprintf("%d. **<@%s>** - niveau %d", i+1, c.DiscordID, xp.Level(c.XP))
			} else {
				msg += fmt.Sprintf("%d. **<@%s>** - niveau %d\n", i+1, c.DiscordID, xp.Level(c.XP))
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
