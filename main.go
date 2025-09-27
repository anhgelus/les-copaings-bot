package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"
	"regexp"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/commands"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/joho/godotenv"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
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
		slog.Error("error while loading .env file", "error", err)
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
		bot.Session,
		*event.InteractionCreate,
		*interaction.MessageComponentData,
		[]string, *cmd.ResponseBuilder,
	),
	pattern string,
) {
	compiledPattern := regexp.MustCompile(pattern)
	b.AddHandler(func(s bot.Session, i *event.InteractionCreate) {
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
		bot.Session,
		*event.InteractionCreate,
		*interaction.ModalSubmitData,
		[]string,
		*cmd.ResponseBuilder,
	),
	pattern string,
) {
	compiledPattern := regexp.MustCompile(pattern)
	b.AddHandler(func(s bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionModalSubmit {
			return
		}

		data := i.ModalSubmitData()
		content, _ := json.Marshal(data)
		s.Logger().Debug(string(content))
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

	b := gokord.Bot{
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
				d = 3 * exp.DebugFactor * time.Second
			}

			user.PeriodicReducer(dg)

			stopPeriodicReducer = gokord.NewTimer(d, func(stop chan<- interface{}) {
				dg.Logger().Debug("periodic reducer")
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
	b.HandleMessageComponent(commands.ConfigMessageComponent, commands.OpenConfig)
	// xp role related
	b.HandleMessageComponent(config.HandleXpRole, config.ModifyXpRole)
	b.HandleMessageComponent(config.HandleXpRoleNew, config.XpRoleNew)
	b.HandleModal(config.HandleXpRoleAdd, config.XpRoleAdd)
	handleDynamicMessageComponent(&b, config.HandleXpRoleEdit, config.XpRoleEditPattern)
	handleDynamicMessageComponent(&b, config.HandleXpRoleEditRole, config.XpRoleEditRolePattern)
	handleDynamicMessageComponent(&b, config.HandleXpRoleEditLevelStart, config.XpRoleEditLevelStartPattern)
	handleDynamicModalComponent(&b, config.HandleXpRoleEditLevel, config.XpRoleEditLevelPattern)
	handleDynamicMessageComponent(&b, config.HandleXpRoleDel, config.XpRoleDel)
	// channel related
	b.HandleMessageComponent(func(s bot.Session, i *event.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
		if config.HandleModifyFallbackChannel(s, i, data, resp) {
			commands.ConfigMessageComponent(s, i, data, resp)
		}
	}, config.ModifyFallbackChannel)
	b.HandleMessageComponent(func(s bot.Session, i *event.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
		if config.HandleModifyDisChannel(s, i, data, resp) {
			commands.ConfigMessageComponent(s, i, data, resp)
		}
	}, config.ModifyDisChannel)
	// reduce related
	b.HandleMessageComponent(config.HandleModifyPeriodicReduceCommand, config.ModifyTimeReduce)
	b.HandleModal(func(s bot.Session, i *event.InteractionCreate, data *interaction.ModalSubmitData, resp *cmd.ResponseBuilder) {
		if config.HandleTimeReduceSet(s, i, data, resp) {
			commands.ConfigModal(s, i, data, resp)
		}
	}, config.TimeReduceSet)

	// xp handlers
	b.AddHandler(OnMessage)
	b.AddHandler(OnVoiceUpdate)
	b.AddHandler(OnLeave)

	b.Start()

	if stopPeriodicReducer != nil {
		stopPeriodicReducer <- true
	}
}
