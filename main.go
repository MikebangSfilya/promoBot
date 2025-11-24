package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"promo-bot/db/repo"
	"promo-bot/handlers"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/app"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/metrics"
	"github.com/kozalosev/goSadTgBot/server"
	"github.com/kozalosev/goSadTgBot/storage"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/sirupsen/logrus"
)

var (
	locpool            = loc.NewPool("en")
	supportedLanguages = []string{"en", "ru"}
)

func main() {
	// the application is listening for the SIGTERM signal to exit
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	metrics.AddHttpHandlerForMetrics()

	srv := server.Start(os.Getenv("APP_PORT"))

	stateStorage, db := establishConnections(ctx)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		panic(err)
	}

	api := base.NewBotAPI(bot)

	debugMode := os.Getenv("DEBUG")
	bot.Debug = strings.ToLower(debugMode) == "true" || debugMode == "1"

	appenv := &base.ApplicationEnv{
		Bot:      api,
		Database: db,
		Ctx:      ctx,
	}

	messageHandlers, callbackHandlers := initHandlers(appenv, stateStorage)
	api.SetCommands(locpool, supportedLanguages, base.ConvertHandlersToCommands(messageHandlers))

	appParams := &app.Params{
		Ctx:              ctx,
		MessageHandlers:  messageHandlers,
		CallbackHandlers: callbackHandlers,
		Settings:         repo.NewUserService(appenv),
		LangPool:         locpool,
		API:              api,
		StateStorage:     stateStorage,
		DB:               db,
	}

	if wasPopulated := wizard.PopulateWizardDescriptors(messageHandlers); !wasPopulated {
		log.Fatal()
	}

	var (
		wg         sync.WaitGroup
		wasStopped bool
	)
	//Graceful shot виидмо
	if bot.Debug {
		if _, err := bot.Request(tgbotapi.DeleteWebhookConfig{}); err != nil {
			panic(err)
		}

		updateConfig := tgbotapi.UpdateConfig{Offset: 0, Timeout: 30, AllowedUpdates: []string{
			tgbotapi.UpdateTypeMessage,
			tgbotapi.UpdateTypeInlineQuery,
			tgbotapi.UpdateTypeCallbackQuery,
			tgbotapi.UpdateTypeChosenInlineResult,
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
	shutdown(stateStorage, db)
}

func establishConnections(ctx context.Context) (stateStorage wizard.StateStorage, db *pgxpool.Pool) {
	commandStateTTL, err := time.ParseDuration(os.Getenv("COMMAND_STATE_TTL"))
	if err != nil {
		panic(err)
	}
	stateStorage = wizard.ConnectToRedis(ctx, commandStateTTL, &redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	dbName := os.Getenv("POSTGRES_DB")
	dbConfig := storage.NewDatabaseConfig(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		dbName)
	db = storage.ConnectToDatabase(ctx, dbConfig)
	// storage.RunMigrations(dbConfig, os.Getenv("MIGRATIONS_REPO"))
	metrics.RegisterMetricsForPgxPoolStat(db, dbName)
	return
}

func initHandlers(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) (messageHandlers []base.MessageHandler, callbackHandlers []base.CallbackHandler) {
	languageHandler := handlers.NewLanguageHandler(appEnv, stateStorage)
	messageHandlers = []base.MessageHandler{
		// handlers.NewHelpHandler(languageHandler),
		languageHandler,
		handlers.NewPromoHanlder(appEnv, stateStorage),
		// handlers.NewNotPrivateChatFallbackHandler(stateStorage),
	}
	callbackHandlers = []base.CallbackHandler{
		// handlers.NewRevokeCallbackHandler(appEnv),
	}
	// metrics.RegisterMessageHandlerCounters(messageHandlers...)
	return
}

func shutdown(stateStorage wizard.StateStorage, db *pgxpool.Pool) {
	db.Close()
	if err := stateStorage.Close(); err != nil {
		logrus.WithField(logconst.FieldFunc, "shutdown").
			WithField(logconst.FieldCalledObject, "StateStorage").
			WithField(logconst.FieldCalledMethod, "Close").
			Error(err)
	}
}
