package main

import (
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
)

var token string

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	err := gokord.SetupConfigs([]*gokord.ConfigInfo{})
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&xp.Copaing{})
	if err != nil {
		panic(err)
	}

	rankCmd := gokord.NewCommand("rank", "Affiche le niveau d'une personne").
		HasOption().
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"copaing",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank).
		ToCmd()

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot 2.0",
				Url:     "",
			},
		},
		Commands: []*gokord.Cmd{
			rankCmd,
		},
		AfterInit: afterInit,
	}
	bot.Start()
}

func afterInit(dg *discordgo.Session) {
	dg.AddHandler(xp.OnMessage)
}
