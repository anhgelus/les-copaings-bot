package main

import (
	_ "embed"
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/user"
	"github.com/bwmarrin/discordgo"
	"time"
)

var (
	token string
	//go:embed updates.json
	updatesData []byte
	Version     = gokord.Version{
		Major: 3,
		Minor: 0,
		Patch: 0,
	}

	stopPeriodicReducer chan<- interface{}
)

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	err := gokord.SetupConfigs(nil, []*gokord.ConfigInfo{})
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&user.Copaing{}, &config.GuildConfig{}, &config.XpRole{})
	if err != nil {
		panic(err)
	}

	adm := gokord.AdminPermission

	rankCmd := gokord.NewCommand("rank", "Affiche le niveau d'un user").
		HasOption().
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"user",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank)

	configCmd := gokord.NewCommand("config", "Modifie la config").
		ContainsSub().
		AddSub(
			gokord.NewCommand("show", "Affiche la config").SetHandler(commands.ConfigShow),
		).
		AddSub(
			gokord.NewCommand("exp", "Modifie l'exp").
				HasOption().
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionString,
					"type",
					"Type d'action à effectuer",
				).
					AddChoice(gokord.NewChoice("Ajouter", "add")).
					AddChoice(gokord.NewChoice("Supprimer", "del")).
					AddChoice(gokord.NewChoice("Modifier", "edit")).IsRequired(),
				).
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionInteger,
					"level",
					"Niveau du rôle",
				).IsRequired()).
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionRole,
					"role",
					"Rôle",
				).IsRequired()).
				SetHandler(commands.ConfigXP),
		).
		AddSub(
			gokord.NewCommand("disabled-channels", "Modifie les salons désactivés").
				HasOption().
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionString,
					"type",
					"Type d'action à effectuer",
				).
					AddChoice(gokord.NewChoice("Désactiver le salon", "add")).
					AddChoice(gokord.NewChoice("Activer le salon", "del")).IsRequired(),
				).
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionChannel,
					"channel",
					"Salon à modifier",
				).IsRequired()).
				SetHandler(commands.ConfigChannel),
		).
		AddSub(
			gokord.NewCommand("fallback-channel", "Modifie le salon textuel par défaut").
				HasOption().
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionChannel,
					"channel",
					"Salon textuel par défaut",
				).IsRequired()).
				SetHandler(commands.ConfigFallbackChannel),
		).SetPermission(&adm)

	topCmd := gokord.NewCommand("top", "Copaings les plus actifs").
		HasOption().
		SetHandler(commands.Top)

	resetCmd := gokord.NewCommand("reset", "Reset l'exp").
		HasOption().
		SetHandler(commands.Reset).
		SetPermission(&adm)

	resetUserCmd := gokord.NewCommand("reset-user", "Reset l'exp d'un utilisation").
		HasOption().
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"user",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser).
		SetPermission(&adm)

	creditsCmd := gokord.NewCommand("credits", "Crédits").
		HasOption().
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
		Commands: []*gokord.GeneralCommand{
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
	}
	bot.Start()

	if stopPeriodicReducer != nil {
		stopPeriodicReducer <- true
	}

	config.CloseRedisClient()
}

func afterInit(dg *discordgo.Session) {
	// handlers
	dg.AddHandler(OnMessage)
	dg.AddHandler(OnVoiceUpdate)
	dg.AddHandler(OnLeave)

	// setup timer for periodic reducer
	d := 24 * time.Hour
	if gokord.Debug {
		// reduce time for debug
		d = time.Minute
	}
	stopPeriodicReducer = utils.NewTimer(d, func(stop chan<- interface{}) {
		user.PeriodicReducer(dg)
	})
}
