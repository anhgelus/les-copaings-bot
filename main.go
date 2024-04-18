package main

import (
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
	"time"
)

var token string

const (
	Version = "2.2.1" // git version: 0.2.1 (it's the v2 of the bot)
)

func init() {
	flag.StringVar(&token, "token", "", "token of the bot")
	flag.Parse()
}

func main() {
	err := gokord.SetupConfigs([]*gokord.ConfigInfo{})
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&xp.Copaing{}, &config.GuildConfig{}, &config.XpRole{})
	if err != nil {
		panic(err)
	}

	rankCmd := gokord.NewCommand("rank", "Affiche le niveau d'un copaing").
		HasOption().
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
					AddChoice(gokord.NewChoice("Désactiver", "add")).
					AddChoice(gokord.NewChoice("Activer", "del")).IsRequired(),
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
		).SetPermission(gokord.AdminPermission)

	topCmd := gokord.NewCommand("top", "Copaings les plus actifs").
		HasOption().
		SetHandler(commands.Top)

	resetCmd := gokord.NewCommand("reset", "Reset l'xp").
		HasOption().
		SetHandler(commands.Reset).
		SetPermission(gokord.AdminPermission)

	resetUserCmd := gokord.NewCommand("reset-user", "Reset l'xp d'un utilisation").
		HasOption().
		AddOption(gokord.NewOption(
			discordgo.ApplicationCommandOptionUser,
			"copaing",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser).
		SetPermission(gokord.AdminPermission)

	creditsCmd := gokord.NewCommand("credits", "Crédits").
		HasOption().
		SetHandler(commands.Credits)

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.WatchStatus,
				Content: "Les Copaings",
			},
			{
				Type:    gokord.GameStatus,
				Content: "dev par @anhgelus",
			},
			{
				Type:    gokord.ListeningStatus,
				Content: "http 418, I'm a tea pot",
			},
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot " + Version,
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
		AfterInit: afterInit,
	}
	bot.Start()

	xp.CloseRedisClient()
}

func afterInit(dg *discordgo.Session) {
	// handlers
	dg.AddHandler(xp.OnMessage)
	dg.AddHandler(xp.OnVoiceUpdate)
	dg.AddHandler(xp.OnLeave)

	// setup timer for periodic reducer
	d := 24 * time.Hour
	if gokord.Debug {
		// reduce time for debug
		d = time.Minute
	}
	utils.NewTimer(d, func(stop chan struct{}) {
		xp.PeriodicReducer(dg)
	})
}
