package main

import (
	_ "embed"
	"errors"
	"flag"
	"github.com/anhgelus/gokord"
	cmd "github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"os"
	"time"
)

var (
	token string
	//go:embed updates.json
	updatesData []byte
	Version     = gokord.Version{
		Major: 3,
		Minor: 1,
		Patch: 3,
	}

	stopPeriodicReducer chan<- interface{}
)

func init() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Warn("Error while loading .env file", "error", err.Error())
	}
	flag.StringVar(&token, "token", os.Getenv("TOKEN"), "token of the bot")
}

func main() {
	flag.Parse()
	gokord.UseRedis = false
	err := gokord.SetupConfigs(&Config{}, []*gokord.ConfigInfo{})
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&user.Copaing{}, &config.GuildConfig{}, &config.XpRole{}, &user.CopaingXP{})
	if err != nil {
		panic(err)
	}

	adm := gokord.AdminPermission

	rankCmd := cmd.New("rank", "Affiche le niveau d'un copaing").
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"copaing",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank)

	configCmd := cmd.New("config", "Modifie la config").
		SetPermission(&adm).
		SetHandler(commands.Config)

	topCmd := cmd.New("top", "Copaings les plus actifs").
		SetHandler(commands.Top)

	resetCmd := cmd.New("reset", "Reset l'xp").
		SetHandler(commands.Reset).
		SetPermission(&adm)

	resetUserCmd := cmd.New("reset-user", "Reset l'xp d'un utilisation").
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"user",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser).
		SetPermission(&adm)

	creditsCmd := cmd.New("credits", "Crédits").
		SetHandler(commands.Credits)

	innovations, err := gokord.LoadInnovationFromJson(updatesData)
	if err != nil {
		panic(err)
	}

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.WatchStatus,
				Content: "Les Copaings",
			},
			{
				Type:    gokord.GameStatus,
				Content: "être dev par @anhgelus",
			},
			{
				Type:    gokord.ListeningStatus,
				Content: "http 418, I'm a tea pot",
			},
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot " + Version.String(),
			},
		},
		Commands: []cmd.CommandBuilder{
			rankCmd,
			configCmd,
			topCmd,
			resetCmd,
			resetUserCmd,
			creditsCmd,
		},
		AfterInit:   afterInit,
		Innovations: innovations,
		Version:     &Version,
		Intents: discordgo.IntentsAllWithoutPrivileged |
			discordgo.IntentsMessageContent |
			discordgo.IntentGuildMembers,
	}
	bot.Start()

	if stopPeriodicReducer != nil {
		stopPeriodicReducer <- true
	}
}

func afterInit(dg *discordgo.Session) {
	// handlers
	dg.AddHandler(OnMessage)
	dg.AddHandler(OnVoiceUpdate)
	dg.AddHandler(OnLeave)

	stopPeriodicReducer = gokord.NewTimer(24*time.Hour, func(stop chan<- interface{}) {
		user.PeriodicReducer(dg)
	})

	//interaction: /config
	dg.AddHandler(commands.ConfigXP)
	dg.AddHandler(commands.ConfigXPModal)
}
