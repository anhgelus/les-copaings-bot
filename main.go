package main

import (
	_ "embed"
	"errors"
	"flag"
	"os"
	"time"

	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	token string
	//go:embed updates.json
	updatesData []byte
	Version     = gokord.Version{
		Major: 3,
		Minor: 2,
		Patch: 0,
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
		AfterInit: func(dg *discordgo.Session) {
			d := 24 * time.Hour
			if gokord.Debug {
				d = 24 * time.Second
			}

			user.PeriodicReducer(dg)

			stopPeriodicReducer = gokord.NewTimer(d, func(stop chan<- interface{}) {
				logger.Debug("Periodic reducer")
				user.PeriodicReducer(dg)
			})
		},
		Innovations: innovations,
		Version:     &Version,
		Intents: discordgo.IntentsAllWithoutPrivileged |
			discordgo.IntentsMessageContent |
			discordgo.IntentGuildMembers,
	}

	// interaction: /config
	// xp role related
	bot.HandleMessageComponent(config.HandleModifyXpRole, config.ModifyXpRole)
	bot.HandleMessageComponent(config.HandleXpRoleAddEdit, config.XpRoleAdd)
	bot.HandleMessageComponent(config.HandleXpRoleAddEdit, config.XpRoleEdit)
	bot.HandleMessageComponent(config.HandleXpRoleAddRole, config.XpRoleAddRole)
	bot.HandleMessageComponent(config.HandleXpRoleEditRole, config.XpRoleEditRole)
	bot.HandleMessageComponent(config.HandleXpRoleDel, config.XpRoleDel)
	bot.HandleMessageComponent(config.HandleXpRoleDelRole, config.XpRoleDelRole)
	bot.HandleModal(config.HandleXpRoleLevel, config.XpRoleAddLevel)
	bot.HandleModal(config.HandleXpRoleLevel, config.XpRoleEditLevel)
	// channel related
	bot.HandleMessageComponent(config.HandleModifyFallbackChannel, config.ModifyFallbackChannel)
	bot.HandleMessageComponent(config.HandleFallbackChannelSet, config.FallbackChannelSet)
	bot.HandleMessageComponent(config.HandleModifyDisChannel, config.ModifyDisChannel)
	bot.HandleMessageComponent(config.HandleDisChannel, config.DisChannelAdd)
	bot.HandleMessageComponent(config.HandleDisChannel, config.DisChannelDel)
	bot.HandleMessageComponent(config.HandleDisChannelAddSet, config.DisChannelAddSet)
	bot.HandleMessageComponent(config.HandleDisChannelDelSet, config.DisChannelDelSet)
	// reduce related
	bot.HandleMessageComponent(config.HandleModifyPeriodicReduce, config.ModifyTimeReduce)
	bot.HandleModal(config.HandleTimeReduceSet, config.TimeReduceSet)

	// xp handlers
	bot.AddHandler(OnMessage)
	bot.AddHandler(OnVoiceUpdate)
	bot.AddHandler(OnLeave)

	bot.Start()

	if stopPeriodicReducer != nil {
		stopPeriodicReducer <- true
	}
}
