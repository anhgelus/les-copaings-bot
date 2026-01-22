package main

import (
	"context"
	_ "embed"
	"flag"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.anhgelus.world/anhgelus/les-copaings-bot/commands"
	"git.anhgelus.world/anhgelus/les-copaings-bot/common"
	"git.anhgelus.world/anhgelus/les-copaings-bot/config"
	"git.anhgelus.world/anhgelus/les-copaings-bot/dynamicid"
	"git.anhgelus.world/anhgelus/les-copaings-bot/exp"
	"git.anhgelus.world/anhgelus/les-copaings-bot/rolereact"
	"git.anhgelus.world/anhgelus/les-copaings-bot/user"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
	"golang.org/x/image/font/opentype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
)

const Version = "3.4.0"

// Args
var (
	token   string
	cfgPath string = "config.toml"
	verbose bool
)

// Cancel timers
var (
	stopPeriodicReducer context.CancelFunc
	stopPeriodicSaver   context.CancelFunc
	stopStatus          context.CancelFunc
)

//go:embed assets/inter-variable.ttf
var interTTF []byte

func init() {
	flag.StringVar(&token, "token", os.Getenv("TOKEN"), "token of the bot")
	flag.StringVar(&cfgPath, "config", cfgPath, "config's path")
	flag.BoolVar(&verbose, "v", verbose, "verbose")

	// Use a nicer font
	fontTTF, err := opentype.Parse(interTTF)
	if err != nil {
		panic(err)
	}
	inter := font.Font{Typeface: "Inter"}
	font.DefaultCache.Add([]font.Face{{
		Font: inter,
		Face: fontTTF,
	}})
	plot.DefaultFont = inter
}

func main() {
	flag.Parse()

	cfg, err := common.LoadConfig(cfgPath)
	if err != nil {
		panic(err)
	}

	db, err := cfg.Database.Connect()
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&user.Copaing{}, &config.GuildConfig{}, &config.XpRole{}, &user.CopaingXP{}, &config.RoleReactMessage{}, &config.RoleReact{})
	if err != nil {
		panic(err)
	}

	adm := int64(discord.PermissionManageGuild)

	ctx := user.SetState(context.Background(), user.NewState())
	ctx = common.SetDB(ctx, db)
	ctx = common.SetDebug(ctx, cfg.Debug)
	ctx = common.SetAuthor(ctx, cfg.Author)

	/*rankCmd := cmd.New("rank", "Affiche le niveau d'un copaing").
		AddOption(cmd.NewOption(
			types.CommandOptionUser,
			"copaing",
			"Le niveau du Copaing que vous souhaitez obtenir",
		)).
		SetHandler(commands.Rank(ctx))

	configCmd := cmd.New("config", "Modifie la config").
		SetPermission(&adm).
		SetHandler(commands.ConfigCommand)

	topCmd := cmd.New("top", "Copaings les plus actifs").
		SetHandler(commands.Top(ctx))

	resetCmd := cmd.New("reset", "Reset l'xp").
		SetHandler(commands.Reset).
		SetPermission(&adm)

	resetUserCmd := cmd.New("reset-user", "Reset l'xp d'un utilisation").
		AddOption(cmd.NewOption(
			types.CommandOptionUser,
			"user",
			"Copaing a reset",
		).IsRequired()).
		SetHandler(commands.ResetUser(ctx)).
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
		SetHandler(commands.Stats(ctx))

	rolereactCmd := cmd.New("rolereact", "Envoie un message permettant de récupérer des rôles grâce à des réactions").
		SetPermission(&adm).
		AddOption(cmd.NewOption(
			types.CommandOptionChannel,
			"salon",
			"Destination du message",
		)).
		SetHandler(rolereact.HandleCommand)*/

	logLevel := slog.LevelInfo
	if verbose {
		logLevel = slog.LevelDebug
	}
	dg := gokord.NewWithLogLevel("Bot "+token, logLevel)

	dg.Identify.Intents = discord.IntentsAllWithoutPrivileged |
		discord.IntentsMessageContent |
		discord.IntentGuildMembers

	events := dg.EventManager()
	intrs := dg.InteractionManager()
	events.AddHandler(ready)

	// related to rolereact
	// TEMP BECAUSE (OLD) GOKORD DOES NOT SUPPORT COMMAND MESSAGE
	events.AddHandler(func(ctx context.Context, s bot.Session, e *event.Ready) {
		guildID := ""
		logger := bot.Logger(ctx)
		if common.IsDebug(ctx) {
			gs, err := s.GuildAPI().UserGuilds(1, "", "", false).Do(ctx)
			if err != nil {
				logger.Error("fetching guilds for debug", "error", err)
				return
			} else {
				guildID = gs[0].ID
			}
		}

		handleRolereactionMessageCmd := interaction.Command{
			Type:                     types.CommandMessage,
			Name:                     "Modifier",
			DefaultMemberPermissions: &adm,
		}
		c, err := s.InteractionAPI().
			CommandCreate(s.SessionState().User().ID, guildID, &handleRolereactionMessageCmd).
			Do(ctx)
		if err != nil {
			logger.Error("unable to push rolereaction message command", "error", err)
			return
		}
		logger.Debug("pushed rolereaction message command", "CommandID", c.ID)
	})
	events.AddHandler(func(ct context.Context, s bot.Session, _ *event.Disconnect) {
		user.PeriodicSaver(ctx, s)
	})
	events.AddHandler(func(ctx context.Context, s bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionApplicationCommand {
			return
		}
		cmd := i.Command()
		if cmd.Data.Name == "Modifier" && cmd.Data.CommandType == types.CommandMessage {
			rolereact.HandleModifyCommand(ctx, s, cmd)
		}
	})
	events.AddHandler(rolereact.HandleReactionAdd)
	events.AddHandler(rolereact.HandleReactionRemove)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleModifyComponent, rolereact.OpenMessage)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleApplyMessage, rolereact.ApplyMessage)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleResetMessage, rolereact.ResetMessage)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleStartSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicModalComponent(events, rolereact.HandleSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleNewRole, rolereact.NewRole)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleOpenRole, rolereact.OpenRole)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleSetRole, rolereact.SetRoleRoleID)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleSetReaction, rolereact.SetRoleReaction)
	dynamicid.HandleDynamicMessageComponent(events, rolereact.HandleDelRole, rolereact.DelRole)

	// interaction: /config
	intrs.HandleMessageComponent(commands.OpenConfig, commands.ConfigMessageComponent)
	// xp role related
	intrs.HandleMessageComponent(config.ModifyXpRole, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) {
		config.HandleXpRole(ctx, dg, i.Interaction)
	})
	intrs.HandleMessageComponent(config.XpRoleNew, config.HandleXpRoleNew)
	intrs.HandleModalSubmit(config.XpRoleAdd, config.HandleXpRoleAdd)
	dynamicid.HandleDynamicMessageComponent(events, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *config.XpRoleId) {
		config.HandleXpRoleEdit(ctx, dg, i.Interaction, params)
	}, config.XpRoleEdit)
	dynamicid.HandleDynamicMessageComponent(events, config.HandleXpRoleEditRole, config.XpRoleEditRole)
	dynamicid.HandleDynamicMessageComponent(events, config.HandleXpRoleEditLevelStart, config.XpRoleEditLevelStart)
	dynamicid.HandleDynamicModalComponent(events, config.HandleXpRoleEditLevel, config.XpRoleEditLevel)
	dynamicid.HandleDynamicMessageComponent(events, config.HandleXpRoleDel, config.XpRoleDel)
	// channel related
	intrs.HandleMessageComponent(config.ModifyFallbackChannel, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) {
		if config.HandleModifyFallbackChannel(ctx, dg, i) {
			commands.ConfigMessageComponent(ctx, dg, i)
		}
	})
	intrs.HandleMessageComponent(config.ModifyDisChannel, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) {
		if config.HandleModifyDisChannel(ctx, dg, i) {
			commands.ConfigMessageComponent(ctx, dg, i)
		}
	})
	// reduce related
	intrs.HandleMessageComponent(config.ModifyTimeReduce, config.HandleModifyPeriodicReduceCommand)
	intrs.HandleModalSubmit(config.TimeReduceSet, func(ctx context.Context, dg bot.Session, i *interaction.ModalSubmit) {
		if config.HandleTimeReduceSet(ctx, dg, i) {
			commands.ConfigModal(ctx, dg, i)
		}
	})

	// xp handlers
	events.AddHandler(OnMessage)
	events.AddHandler(OnVoiceUpdate)
	events.AddHandler(OnLeave)

	err = dg.Open(ctx)
	if err != nil {
		panic(err)
	}

	setupTimers(ctx, dg)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Logger().Info("stopping")

	err = dg.Close(ctx)
	if err != nil {
		panic(err)
	}

	if stopPeriodicReducer != nil {
		stopPeriodicReducer()
	}

	if stopPeriodicSaver != nil {
		stopPeriodicSaver()
	}
}

func setupTimers(ctx context.Context, dg *gokord.Session) {
	d := 24 * time.Hour
	debug := common.IsDebug(ctx)
	if debug {
		d = 3 * exp.DebugFactor * time.Second
	}
	d2 := 30 * time.Minute
	if debug {
		d2 = 1 * exp.DebugFactor * time.Second
	}

	// because logger was never set in this context
	ctx = bot.SetLogger(ctx, dg.Logger())
	user.PeriodicReducer(ctx, dg)

	stopPeriodicReducer = common.NewTimer(ctx, d, func(ctx context.Context, _ context.CancelFunc) {
		user.PeriodicReducer(ctx, dg)
	})
	stopPeriodicSaver = common.NewTimer(ctx, d2, func(ctx context.Context, _ context.CancelFunc) {
		user.PeriodicSaver(ctx, dg)
	})
}

var statuses = []func(context.Context, bot.Session) error{
	func(ctx context.Context, dg bot.Session) error {
		return dg.BotAPI().UpdateGameStatus(ctx, 0, "être dev par @anhgelus")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.BotAPI().UpdateWatchStatus(ctx, 0, "Les Copaings")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.BotAPI().UpdateListeningStatus(ctx, "http 418, I'm a tea pot")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.BotAPI().UpdateGameStatus(ctx, 0, "Les Copaings Bot v"+Version)
	},
}

func ready(ctx context.Context, dg bot.Session, _ *event.Ready) {
	logger := bot.Logger(ctx)
	logger.Info("bot started", "as", dg.SessionState().User().Username)
	now := time.Now().Unix()
	rd := rand.New(rand.NewPCG(uint64(now), uint64(len(statuses))))
	stopStatus = common.NewTimer(ctx, 30*time.Second, func(ctx context.Context, _ context.CancelFunc) {
		rnd := rd.UintN(uint(len(statuses)))
		err := statuses[rnd](ctx, dg)
		if err != nil {
			bot.Logger(ctx).Error("updating status", "error", err)
		}
	})
}
