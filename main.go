package main

import (
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
)

var token string

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	err := gokord.SetupConfigs([]*gokord.ConfigInfo{})
	if err != nil {
		utils.SendError(err)
	}

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot 2.0",
				Url:     "",
			},
		},
		Commands: nil,
		Handlers: nil,
	}
	bot.Start()
}
