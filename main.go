package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	cfg "github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/service/promo"

	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/handlers"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/app"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/metrics"
	"github.com/kozalosev/goSadTgBot/server"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
)

var (
	locpool            = loc.NewPool("ru")
	supportedLanguages = []string{"en", "ru"}
)

func main() {
	DevLvl := os.Getenv(cfg.EnvDevLevel)
	if DevLvl == "" {
		DevLvl = "local"
	}
	log := initLogger(DevLvl)
	log.Info("starting up", "devLvl", DevLvl)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics.AddHttpHandlerForMetrics()

	srv := server.Start(os.Getenv(cfg.EnvAppPort))

	stateStorage, db := establishConnections(ctx)
	txManager := repo.NewTxManager(db)

	bot, err := tgbotapi.NewBotAPI(os.Getenv(cfg.EnvAPIToken))
	if err != nil {
		log.Error("failed to create bot API",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "tgbotapi.NewBotAPI")))
		os.Exit(1)
	}

	botAPI := base.NewBotAPI(bot)

	debugMode := os.Getenv(cfg.EnvDebug)
	bot.Debug = strings.ToLower(debugMode) == "true" || debugMode == "1"

	filePath := os.Getenv(cfg.EnvAuditLogsDir)
	log.Debug("filePath", "path", filePath)

	auditStorage, err := audit.NewFileStorage(filePath)
	if err != nil {
		log.Error("failed to initialize audit storage",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "InitAudit")))
		os.Exit(1)
	}

	appEnv := &base.ApplicationEnv{
		Bot:      botAPI,
		Database: db,
		Ctx:      ctx,
	}

	http.Handle("/promo/generate", restApiStart(appEnv, auditStorage, txManager))
	messageHandlers, callbackHandlers := initHandlers(appEnv, stateStorage, auditStorage, txManager)
	botAPI.SetCommands(locpool, supportedLanguages, base.ConvertHandlersToCommands(messageHandlers))

	usersConfigPath := os.Getenv(cfg.EnvUsersConfigFile)
	usersConfig, err := cfg.NewUsersConfig(usersConfigPath)
	if err != nil {
		slog.Error("failed to read users configuration file",
			slog.Group("error",
				"message", err.Error(),
				"component", "config.NewUsersConfig",
				"path", usersConfigPath))
		os.Exit(1)
	}

	appParams := &app.Params{
		Ctx:              ctx,
		MessageHandlers:  messageHandlers,
		CallbackHandlers: callbackHandlers,
		Settings:         usersConfig,
		LangPool:         locpool,
		API:              botAPI,
		StateStorage:     stateStorage,
		DB:               db,
	}

	if wasPopulated := wizard.PopulateWizardDescriptors(messageHandlers); !wasPopulated {
		slog.Error("failed to populate wizard descriptors",
			slog.Group("error",
				slog.String("message", "wizard initialization failed"),
				slog.String("component", "wizard.PopulateWizardDescriptors")))
		os.Exit(1)
	}

	var (
		wg         sync.WaitGroup
		wasStopped bool
	)

	if bot.Debug {
		if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			slog.Error("failed to delete webhook",
				slog.Group("error",
					slog.String("message", err.Error()),
					slog.String("component", "bot.Request.DeleteWebhook")))
			os.Exit(1)
		}

		updateConfig := tgbotapi.UpdateConfig{Offset: 0, Timeout: 30, AllowedUpdates: []string{
			tgbotapi.UpdateTypeMessage,
			tgbotapi.UpdateTypeCallbackQuery,
		}}
		updates := bot.GetUpdatesChan(updateConfig)

		for upd := range updates {
			select {
			case <-ctx.Done():
				if !wasStopped {
					bot.StopReceivingUpdates()
					wasStopped = true
				}
			default:
			}
			app.HandleUpdate(appParams, &wg, &upd)
		}
	} else {
		server.AddHttpHandlerForWebhook(bot, appParams, &wg)
		<-ctx.Done()
		server.StopListeningForIncomingRequests(srv)
	}

	wg.Wait()
	shutdown(stateStorage, db, auditStorage)
}

func establishConnections(ctx context.Context) (stateStorage wizard.StateStorage, db *pgxpool.Pool) {
	commandStateTTL, err := time.ParseDuration(os.Getenv(cfg.EnvCommandStateTTL))
	if err != nil {
		slog.Error("failed to parse COMMAND_STATE_TTL",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("value", os.Getenv(cfg.EnvCommandStateTTL))))
		os.Exit(1)
	}
	stateStorage = wizard.ConnectToRedis(ctx, commandStateTTL, &redis.Options{
		Addr:     os.Getenv(cfg.EnvRedisHost) + ":" + os.Getenv(cfg.EnvRedisPort),
		Password: os.Getenv(cfg.EnvRedisPassword),
		DB:       0,
	})
	dbName := os.Getenv(cfg.EnvPostgresDB)
	dbConfig := storage.NewDatabaseConfig(
		os.Getenv(cfg.EnvPostgresHost),
		os.Getenv(cfg.EnvPostgresPort),
		os.Getenv(cfg.EnvPostgresUser),
		os.Getenv(cfg.EnvPostgresPassword),
		dbName)
	db = storage.ConnectToDatabase(ctx, dbConfig)
	metrics.RegisterMetricsForPgxPoolStat(db, dbName)
	return
}

func initHandlers(
	appEnv *base.ApplicationEnv,
	stateStorage wizard.StateStorage,
	auditStorage *audit.FileStorage,
	tx *repo.TxManager) (
	messageHandlers []base.MessageHandler,
	callbackHandlers []base.CallbackHandler) {

	promoRepo := repo.NewPromo(appEnv)
	service := promo.NewSaveService(promoRepo, auditStorage, tx)
	promoHandler := handlers.NewPromoHandler(appEnv, stateStorage, service)
	messageHandlers = []base.MessageHandler{
		handlers.NewGetHandler(appEnv, promoRepo),
		promoHandler,
		handlers.NewStats(appEnv, stateStorage, service),
	}
	callbackHandlers = []base.CallbackHandler{}
	metrics.RegisterMessageHandlerCounters(messageHandlers...)
	return
}

func shutdown(stateStorage wizard.StateStorage, db *pgxpool.Pool, auditStorage *audit.FileStorage) {
	db.Close()
	if err := stateStorage.Close(); err != nil {
		slog.Error("failed to close state storage",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "StateStorage.Close")))
	}
	if err := auditStorage.Close(); err != nil {
		slog.Error("failed to close audit storage",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "AuditStorage.Close")))
	}
}

func getLogLevel(env string) slog.Level {
	switch strings.ToLower(env) {
	case "local":
		return slog.LevelDebug
	case "dev":
		return slog.LevelInfo
	case "prod":
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func initLogger(env string) *slog.Logger {
	lvl := getLogLevel(env)
	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if env == "local" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func restApiStart(appEnv *base.ApplicationEnv, auditS *audit.FileStorage, manager *repo.TxManager) http.Handler {
	promoRepo := repo.NewPromo(appEnv)
	service := promo.NewSaveService(promoRepo, auditS, manager)
	restHandler := handlers.NewOneTimePromoHandler(service)
	return restHandler.GeneratePromo()
}
