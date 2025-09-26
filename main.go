package main

import (
	_ "embed"
	"errors"
	"flag"
	"os"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/commands"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/rolereact"
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

func main() {
	flag.Parse()
	gokord.UseRedis = false
	err := gokord.SetupConfigs(&Config{}, []*gokord.ConfigInfo{})
	if err != nil {
		panic(err)
	}

	err = gokord.DB.AutoMigrate(&user.Copaing{}, &config.GuildConfig{}, &config.XpRole{}, &user.CopaingXP{}, &config.RoleReactMessage{}, &config.RoleReact{})
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

	rolereactCmd := cmd.New("rolereact", "Envoie un message permettant de récupérer des rôles grâce à des réactions").
		SetPermission(&adm).
		AddOption(cmd.NewOption(
			types.CommandOptionChannel,
			"salon",
			"Destination du message",
		)).
		SetHandler(rolereact.HandleCommand)

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
			rolereactCmd,
		},
		AfterInit: func(dg *discordgo.Session) {
			d := 24 * time.Hour
			if gokord.Debug {
				d = 3 * exp.DebugFactor * time.Second
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

	// related to rolereact
	bot.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		var guildID string
		gs, err := s.GuildAPI().UserGuilds(1, "", "", false)
		if err != nil {
			s.LogError(err, "fetching guilds for debug")
			return
		} else {
			guildID = gs[0].ID
		}

		handleRolereactionMessageCmd := interaction.Command{
			Type:                     types.CommandMessage,
			Name:                     "Modifier",
			DefaultMemberPermissions: &adm,
		}
		c, err := s.InteractionAPI().CommandCreate(s.State.User.ID, guildID, &handleRolereactionMessageCmd)
		if err != nil {
			s.LogError(err, "unable to push rolereaction message command")
			return
		}
		s.LogDebug("pushed rolereaction message command, commandid %s", c.ID)
	})
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.LogDebug("Handler successfuly called 1")
		if i.Type != types.InteractionApplicationCommand {
			return
		}
		data := i.CommandData()
		s.LogDebug("Handler successfuly called")
		if "Modifier" == data.Name {
			resp := cmd.NewResponseBuilder(s, i)
			rolereact.HandleModifyCommand(s, i, data, resp)
		}
	})
	bot.AddHandler(rolereact.HandleReactionAdd)
	bot.AddHandler(rolereact.HandleReactionRemove)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleModifyComponent, rolereact.OpenMessage)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleApplyMessage, rolereact.ApplyMessage)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleResetMessage, rolereact.ResetMessage)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleStartSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicModalComponent(&bot, rolereact.HandleSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleNewRole, rolereact.NewRole)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleOpenRole, rolereact.OpenRole)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleSetRole, rolereact.SetRoleRoleID)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleSetReaction, rolereact.SetRoleReaction)
	dynamicid.HandleDynamicMessageComponent(&bot, rolereact.HandleDelRole, rolereact.DelRole)

	// interaction: /config
	bot.HandleMessageComponent(commands.ConfigMessageComponent, commands.OpenConfig)
	// xp role related
	bot.HandleMessageComponent(config.HandleXpRole, config.ModifyXpRole)
	bot.HandleMessageComponent(config.HandleXpRoleNew, config.XpRoleNew)
	bot.HandleModal(config.HandleXpRoleAdd, config.XpRoleAdd)
	dynamicid.HandleDynamicMessageComponent(&bot, config.HandleXpRoleEdit, config.XpRoleEdit)
	dynamicid.HandleDynamicMessageComponent(&bot, config.HandleXpRoleEdit, config.XpRoleEdit)
	dynamicid.HandleDynamicMessageComponent(&bot, config.HandleXpRoleEditRole, config.XpRoleEditRole)
	dynamicid.HandleDynamicMessageComponent(&bot, config.HandleXpRoleEditLevelStart, config.XpRoleEditLevelStart)
	dynamicid.HandleDynamicModalComponent(&bot, config.HandleXpRoleEditLevel, config.XpRoleEditLevel)
	dynamicid.HandleDynamicMessageComponent(&bot, config.HandleXpRoleDel, config.XpRoleDel)
	// channel related
	bot.HandleMessageComponent(func(s *discordgo.Session, i *discordgo.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
		if config.HandleModifyFallbackChannel(s, i, data, resp) {
			commands.ConfigMessageComponent(s, i, data, resp)
		}
	}, config.ModifyFallbackChannel)
	bot.HandleMessageComponent(func(s *discordgo.Session, i *discordgo.InteractionCreate, data *interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
		if config.HandleModifyDisChannel(s, i, data, resp) {
			commands.ConfigMessageComponent(s, i, data, resp)
		}
	}, config.ModifyDisChannel)
	// reduce related
	bot.HandleMessageComponent(config.HandleModifyPeriodicReduceCommand, config.ModifyTimeReduce)
	bot.HandleModal(func(s *discordgo.Session, i *discordgo.InteractionCreate, data *interaction.ModalSubmitData, resp *cmd.ResponseBuilder) {
		if config.HandleTimeReduceSet(s, i, data, resp) {
			commands.ConfigModal(s, i, data, resp)
		}
	}, config.TimeReduceSet)

	// xp handlers
	bot.AddHandler(OnMessage)
	bot.AddHandler(OnVoiceUpdate)
	bot.AddHandler(OnLeave)

	bot.Start()

	if stopPeriodicReducer != nil {
		stopPeriodicReducer <- true
	}
}
