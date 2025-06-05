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
		Minor: 1,
		Patch: 3,
	}

	stopPeriodicReducer chan<- interface{}
)

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
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

	rankCmd := gokord.NewCommand("rank", "Affiche le niveau d'un copaing").
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"copaing",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank)

	configCmd := gokord.NewCommand("config", "Modifie la config").
		ContainsSub().
		AddSub(
			gokord.NewCommand("show", "Affiche la config").SetHandler(commands.ConfigShow),
		).
		AddSub(
			gokord.NewCommand("xp", "Modifie l'xp").
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
			gokord.NewCommand("period-before-reduce", "Temps avant la perte d'xp (affecte aussi le /top)").
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionInteger,
					"days",
					"Nombre de jours avant la perte d'xp (doit être égal ou plus grand que 30)",
				).IsRequired()).
				SetHandler(commands.ConfigPeriodBeforeReduce),
		).
		AddSub(
			gokord.NewCommand("fallback-channel", "Modifie le salon textuel par défaut").
				AddOption(gokord.NewOption(
					discordgo.ApplicationCommandOptionChannel,
					"channel",
					"Salon textuel par défaut",
				).IsRequired()).
				SetHandler(commands.ConfigFallbackChannel),
		).SetPermission(&adm)

	topCmd := gokord.NewCommand("top", "Copaings les plus actifs").
		SetHandler(commands.Top)

	resetCmd := gokord.NewCommand("reset", "Reset l'xp").
		SetHandler(commands.Reset).
		SetPermission(&adm)

	resetUserCmd := gokord.NewCommand("reset-user", "Reset l'xp d'un utilisation").
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"user",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser).
		SetPermission(&adm)

	creditsCmd := gokord.NewCommand("credits", "Crédits").
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
		Commands: []gokord.CommandBuilder{
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

	stopPeriodicReducer = utils.NewTimer(24*time.Hour, func(stop chan<- interface{}) {
		user.PeriodicReducer(dg)
	})
}
