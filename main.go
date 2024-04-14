package main

import (
	"flag"
	"github.com/anhgelus/gokord"
)

var token string

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	bot := gokord.Bot{
		Token:    token,
		Status:   nil,
		Commands: nil,
		Handlers: nil,
	}
	bot.Start()
}
