package main

import (
	"flag"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/les-copaings-bot/commands"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/xp"
	"github.com/bwmarrin/discordgo"
)

var token string

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
		)

	bot := gokord.Bot{
		Token: token,
		Status: []*gokord.Status{
			{
				Type:    gokord.GameStatus,
				Content: "Les Copaings Bot 2.0",
				Url:     "",
			},
		},
		Commands: []*gokord.GeneralCommand{
			rankCmd,
			configCmd,
		},
		AfterInit: afterInit,
	}
	bot.Start()

	xp.CloseRedisClient()
}

func afterInit(dg *discordgo.Session) {
	dg.AddHandler(xp.OnMessage)
	dg.AddHandler(xp.OnVoiceUpdate)
	dg.AddHandler(xp.OnLeave)
}
