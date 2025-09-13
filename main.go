package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"regexp"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/commands"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/joho/godotenv"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/discord"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/interaction"
	"github.com/nyttikord/gokord/logger"
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
		logger.Log(logger.LevelError, 0, "Error while loading .env file: %v", err.Error())
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

func handleDynamicMessageComponent(
	b *gokord.Bot,
	handler func(
		*discordgo.Session,
		*discordgo.InteractionCreate,
		interaction.MessageComponentData,
		[]string, *cmd.ResponseBuilder,
	),
	pattern string,
) {
	compiledPattern := regexp.MustCompile(pattern)
	b.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != types.InteractionMessageComponent {
			return
		}

		data := i.MessageComponentData()
		parameters := compiledPattern.FindStringSubmatch(data.CustomID)
		if parameters == nil {
			return
		}
		parameters = parameters[1:]
		handler(s, i, data, parameters, cmd.NewResponseBuilder(s, i))
	})
}

func handleDynamicModalComponent(
	b *gokord.Bot,
	handler func(
		*discordgo.Session,
		*discordgo.InteractionCreate,
		interaction.ModalSubmitData,
		[]string,
		*cmd.ResponseBuilder,
	),
	pattern string,
) {
	compiledPattern := regexp.MustCompile(pattern)
	b.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != types.InteractionModalSubmit {
			return
		}

		data := i.ModalSubmitData()
		content, _ := json.Marshal(data)
		s.LogDebug(string(content))
		parameters := compiledPattern.FindStringSubmatch(data.CustomID)
		if parameters == nil {
			return
		}
		parameters = parameters[1:]
		handler(s, i, data, parameters, cmd.NewResponseBuilder(s, i))
	})
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
			types.CommandOptionUser,
			"copaing",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank)

	configCmd := cmd.New("config", "Modifie la config").
		SetPermission(&adm).
		SetHandler(commands.ConfigCommand)

	topCmd := cmd.New("top", "Copaings les plus actifs").
		SetHandler(commands.Top)

	resetCmd := cmd.New("reset", "Reset l'xp").
		SetHandler(commands.Reset).
		SetPermission(&adm)

	resetUserCmd := cmd.New("reset-user", "Reset l'xp d'un utilisation").
		AddOption(cmd.NewOption(
			types.CommandOptionUser,
			"user",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser).
		SetPermission(&adm)

	creditsCmd := cmd.New("credits", "Crédits").
		SetHandler(commands.Credits)

	statsCmd := cmd.New("stats", "Affiche des stats :D").
		AddOption(cmd.NewOption(
			types.CommandOptionInteger,
			"days",
			"Nombre de jours à afficher dans le graphique",
		)).
		AddOption(cmd.NewOption(
			types.CommandOptionUser,
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
				dg.LogDebug("Periodic reducer")
				user.PeriodicReducer(dg)
			})
		},
		Innovations: innovations,
		Version:     &Version,
		Intents: discord.IntentsAllWithoutPrivileged |
			discord.IntentsMessageContent |
			discord.IntentGuildMembers,
	}

	// interaction: /config
	bot.HandleMessageComponent(func(s *discordgo.Session, i *discordgo.InteractionCreate, data interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
		if len(data.Values) != 1 {
			bot.LogError(errors.New("invalid data values"), "handle config modify, values: %#v", data.Values)
			return
		}
		switch data.Values[0] {
		case config.ModifyXpRole:
			config.HandleXpRole(s, i, data, resp)
		case config.ModifyFallbackChannel:
			config.HandleModifyFallbackChannel(s, i, data, resp)
		case config.ModifyDisChannel:
			config.HandleModifyDisChannel(s, i, data, resp)
		case config.ModifyTimeReduce:
			config.HandleModifyPeriodicReduce(s, i, data, resp)
		default:
			bot.LogError(errors.New("unknown value"), "detecting value %s", data.Values[0])
			return
		}
	}, commands.ConfigModify)
	bot.HandleMessageComponent(commands.ConfigMessageComponent, commands.OpenConfig)
	// xp role related
	bot.HandleMessageComponent(config.HandleXpRole, config.ModifyXpRole)
	bot.HandleMessageComponent(config.HandleXpRoleNew, config.XpRoleNew)
	bot.HandleModal(config.HandleXpRoleAdd, config.XpRoleAdd)
	handleDynamicMessageComponent(&bot, config.HandleXpRoleEdit, config.XpRoleEditPattern)
	handleDynamicMessageComponent(&bot, config.HandleXpRoleEditRole, config.XpRoleEditRolePattern)
	handleDynamicMessageComponent(&bot, config.HandleXpRoleEditLevelStart, config.XpRoleEditLevelStartPattern)
	handleDynamicModalComponent(&bot, config.HandleXpRoleEditLevel, config.XpRoleEditLevelPattern)
	handleDynamicMessageComponent(&bot, config.HandleXpRoleDel, config.XpRoleDel)
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
