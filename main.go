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
	deploy  bool   = false
	verbose bool   = false
)

// Cancel timers
var (
	stopPeriodicReducer context.CancelFunc
	stopPeriodicSaver   context.CancelFunc
	stopStatus          context.CancelFunc
)

var finishDeploy context.CancelFunc

//go:embed assets/inter-variable.ttf
var interTTF []byte

func init() {
	flag.StringVar(&token, "token", os.Getenv("TOKEN"), "token of the bot")
	flag.StringVar(&cfgPath, "config", cfgPath, "config's path")
	flag.BoolVar(&verbose, "v", verbose, "verbose")
	flag.BoolVar(&deploy, "deploy", deploy, "deploy commands")

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
	if cfg == nil {
		println("config initialized")
		return
	}

	db, err := cfg.Database.Connect()
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&user.Copaing{}, &config.Guild{}, &config.XpRole{}, &user.CopaingXP{}, &config.RoleReactMessage{}, &config.RoleReact{})
	if err != nil {
		panic(err)
	}

	ctx := user.SetState(context.Background(), user.NewState(db))
	ctx = common.SetDB(ctx, db)
	ctx = common.SetDebug(ctx, cfg.Debug)
	ctx = common.SetAuthor(ctx, cfg.Author)

	logLevel := slog.LevelInfo
	if verbose || cfg.Debug {
		logLevel = slog.LevelDebug
	}
	dg := gokord.NewWithLogLevel("Bot "+token, logLevel)

	dg.Identify.Intents = discord.IntentsAllWithoutPrivileged |
		discord.IntentsMessageContent |
		discord.IntentGuildMembers

	events := dg.EventManager()
	intrs := dg.InteractionManager()
	// handle deploy
	if deploy {
		var ctx2 context.Context
		ctx2, finishDeploy = context.WithCancel(ctx)
		events.AddHandler(deploying)

		err = dg.Open(ctx)
		if err != nil {
			panic(err)
		}

		dg.Logger().Info("waiting deployment")
		<-ctx2.Done()

		dg.Logger().Info("commands deployed, closing")
		err = dg.Close(ctx)
		if err != nil {
			panic(err)
		}
		return
	}

	events.AddHandler(ready)
	events.AddHandler(func(ct context.Context, dg bot.Session, _ *event.Disconnect) {
		user.PeriodicSaver(ctx, dg)
	})
	events.AddHandler(rolereact.HandleReactionAdd)
	events.AddHandler(rolereact.HandleReactionRemove)

	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleModifyComponent, rolereact.OpenMessage)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleApplyMessage, rolereact.ApplyMessage)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleResetMessage, rolereact.ResetMessage)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleStartSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicModalComponent(intrs, rolereact.HandleSetNote, rolereact.SetNote)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleNewRole, rolereact.NewRole)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleOpenRole, rolereact.OpenRole)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleSetRole, rolereact.SetRoleRoleID)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleSetReaction, rolereact.SetRoleReaction)
	dynamicid.HandleDynamicMessageComponent(intrs, rolereact.HandleDelRole, rolereact.DelRole)

	// commands
	intrs.HandleCommand("rank", commands.Rank)
	intrs.HandleCommand("config", commands.ConfigCommand)
	intrs.HandleCommand("top", commands.Top)
	intrs.HandleCommand("reset", commands.Reset)
	intrs.HandleCommand("reset-user", commands.ResetUser)
	intrs.HandleCommand("credits", commands.Credits)
	intrs.HandleCommand("stats", commands.Stats)
	intrs.HandleCommand("rolereact", rolereact.HandleCommand)
	intrs.HandleCommand("Modifier", rolereact.HandleModifyCommand)

	// interaction: /config
	intrs.HandleMessageComponent(commands.OpenConfig, commands.ConfigMessageComponent)
	// xp role related
	intrs.HandleMessageComponent(config.ModifyXpRole, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) {
		config.HandleXpRole(ctx, dg, i.Interaction)
	})
	intrs.HandleMessageComponent(config.XpRoleNew, config.HandleXpRoleNew)
	intrs.HandleModalSubmit(config.XpRoleAdd, config.HandleXpRoleAdd)
	dynamicid.HandleDynamicMessageComponent(intrs, func(ctx context.Context, dg bot.Session, i *interaction.MessageComponent, params *config.XpRoleId) {
		config.HandleXpRoleEdit(ctx, dg, i.Interaction, params)
	}, config.XpRoleEdit)
	dynamicid.HandleDynamicMessageComponent(intrs, config.HandleXpRoleEditRole, config.XpRoleEditRole)
	dynamicid.HandleDynamicMessageComponent(intrs, config.HandleXpRoleEditLevelStart, config.XpRoleEditLevelStart)
	dynamicid.HandleDynamicModalComponent(intrs, config.HandleXpRoleEditLevel, config.XpRoleEditLevel)
	dynamicid.HandleDynamicMessageComponent(intrs, config.HandleXpRoleDel, config.XpRoleDel)
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

	stopPeriodicReducer = common.NewTimer(ctx, d, func(ctx context.Context, _ context.CancelFunc) {
		user.PeriodicReducer(ctx, dg)
	})
	stopPeriodicSaver = common.NewTimer(ctx, d2, func(ctx context.Context, _ context.CancelFunc) {
		user.PeriodicSaver(ctx, dg)
	})
}

var statuses = []func(context.Context, bot.Session) error{
	func(ctx context.Context, dg bot.Session) error {
		return dg.GatewayAPI().UpdateGameStatus(ctx, false, "être dev par @anhgelus")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.GatewayAPI().UpdateWatchStatus(ctx, false, "Les Copaings")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.GatewayAPI().UpdateListeningStatus(ctx, "http 418, I'm a tea pot")
	},
	func(ctx context.Context, dg bot.Session) error {
		return dg.GatewayAPI().UpdateGameStatus(ctx, false, "Les Copaings Bot v"+Version)
	},
}

func ready(ctx context.Context, dg bot.Session, _ *event.Ready) {
	logger := bot.Logger(ctx)
	logger.Info("bot started", "as", dg.SessionState().User().Username, "usage", "main")
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

func deploying(ctx context.Context, dg bot.Session, _ *event.Connect) {
	logger := bot.Logger(ctx)
	logger.Info("bot started", "as", dg.SessionState().User().Username, "usage", "deploy")
	err := commands.Deploy(ctx, dg)
	if err != nil {
		panic(err)
	}
	finishDeploy()
}
