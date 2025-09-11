package main

import (
	_ "embed"
	"errors"
	"flag"
	"os"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/commands"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/logger"
	"github.com/joho/godotenv"
	discordgo "github.com/nyttikord/gokord"
	"golang.org/x/image/font/opentype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
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

//go:embed assets/inter-variable.ttf
var interTTF []byte

func init() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Warn("Error while loading .env file", "error", err.Error())
	}
	flag.StringVar(&token, "token", os.Getenv("TOKEN"), "token of the bot")

	// Use a nicer font
	fontTTF, parseErr := opentype.Parse(interTTF)
	if parseErr != nil {
		panic(err)
	}
	inter := font.Font{Typeface: "Inter"}
	font.DefaultCache.Add(
		[]font.Face{
			{
				Font: inter,
				Face: fontTTF,
			},
		})
	plot.DefaultFont = inter

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

	statsCmd := cmd.New("stats", "Affiche des stats :D").
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionInteger,
			"days",
			"Nombre de jours à afficher dans le graphique",
		)).
		AddOption(cmd.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"user",
			"Utilisateur à inspecter",
		)).
		SetHandler(commands.Stats)

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
			statsCmd,
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
	bot.HandleMessageComponent(func(s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
		if len(data.Values) != 1 {
			logger.Alert("main.go - Handle config modify", "invalid data values", "values", data.Values)
			return
		}
		switch data.Values[0] {
		case config.ModifyXpRole:
			config.HandleModifyXpRole(s, i, data, resp)
		case config.ModifyFallbackChannel:
			config.HandleModifyFallbackChannel(s, i, data, resp)
		case config.ModifyDisChannel:
			config.HandleModifyDisChannel(s, i, data, resp)
		case config.ModifyTimeReduce:
			config.HandleModifyPeriodicReduce(s, i, data, resp)
		default:
			logger.Alert("main.go - Detecting value", "unkown value", "value", data.Values[0])
			return
		}
	}, commands.ConfigModify)
	// xp role related
	bot.HandleMessageComponent(config.HandleXpRoleAddEdit, config.XpRoleAdd)
	bot.HandleMessageComponent(config.HandleXpRoleAddEdit, config.XpRoleEdit)
	bot.HandleMessageComponent(config.HandleXpRoleAddRole, config.XpRoleAddRole)
	bot.HandleMessageComponent(config.HandleXpRoleEditRole, config.XpRoleEditRole)
	bot.HandleMessageComponent(config.HandleXpRoleDel, config.XpRoleDel)
	bot.HandleMessageComponent(config.HandleXpRoleDelRole, config.XpRoleDelRole)
	bot.HandleModal(config.HandleXpRoleLevel, config.XpRoleAddLevel)
	bot.HandleModal(config.HandleXpRoleLevel, config.XpRoleEditLevel)
	// channel related
	bot.HandleMessageComponent(config.HandleFallbackChannelSet, config.FallbackChannelSet)
	bot.HandleMessageComponent(config.HandleDisChannel, config.DisChannelAdd)
	bot.HandleMessageComponent(config.HandleDisChannel, config.DisChannelDel)
	bot.HandleMessageComponent(config.HandleDisChannelAddSet, config.DisChannelAddSet)
	bot.HandleMessageComponent(config.HandleDisChannelDelSet, config.DisChannelDelSet)
	// reduce related
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
