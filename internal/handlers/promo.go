package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/model"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	UnknowCommand       = "commands.default.message.on.command"
	promoFieldsTrPrefix = "commands.promo.fields."

	BadLength   = "BadLength"
	BadCapacity = "BadCapacity"

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
	fieldLength       = "length"
	fieldCapacity     = "capacity"
	fullMsg           = "messages.promo.created_full"

	actionCreate = "actionCreate"
	actionCancel = "actionCancel"
	textToCreate = "textToCreate"

	promoCanceled    = "promoCanceled"
	errToCreatePromo = "errToCreatePromo"
	errNoPermission  = "errNoPermission"
)

type PromoHandler struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage
	auditStorage audit.Storage

	PromoService *repo.Promo
	txManager    *repo.TxManager
}

func NewPromoHandler(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage, storage audit.Storage, txManager *repo.TxManager) *PromoHandler {
	h := &PromoHandler{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		auditStorage: storage,
		PromoService: repo.NewPromo(appEnv),
		txManager:    txManager,
	}
	h.HandlerRefForTrait = h
	return h
}

func (h *PromoHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(h.appEnv, h.stateStorage)
}

func (h *PromoHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(h.action)

	desc.AddField(fieldPromo, promoFieldsTrPrefix+fieldPromo)

	desc.AddField(fieldLength, promoFieldsTrPrefix+fieldLength)

	desc.AddField(fieldCapacity, promoFieldsTrPrefix+fieldCapacity)

	confirm := desc.AddField(fieldConfirmation, textToCreate)
	confirm.InlineKeyboardAnswers = []string{actionCreate, actionCancel}
	return desc
}

func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	const op = "PromoHandler.Handle"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to cast Options to UserOptions",
			slog.Group("error",
				slog.String("message", "type assertion failed")))
		reply("failure")
		return
	}

	if opts.Role != config.Admin {
		reply(errNoPermission)
		return
	}

	promoForm := wizard.NewWizard(h, 4)

	promoForm.AddEmptyField(fieldPromo, wizard.Text)
	promoForm.AddEmptyField(fieldLength, wizard.Text)
	promoForm.AddEmptyField(fieldCapacity, wizard.Text)
	promoForm.AddEmptyField(fieldConfirmation, wizard.Text)

	promoForm.ProcessNextField(reqEnv, msg)
}

func (h *PromoHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	const op = "PromoHandler.action"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqenv, msg)

	opts, ok := reqenv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to get user options",
			slog.Group("error",
				"message", "type assertion failed"))
		reply(errToCreatePromo)
		return
	}

	promoCode := extractPromoInfo(fields, fieldPromo)
	confirmAct := extractPromoInfo(fields, fieldConfirmation)

	lengthExtract := extractPromoInfo(fields, fieldLength)
	length, err := strToInt(lengthExtract)
	if err != nil {
		log.Error("failed to parse length",
			slog.Group("error",
				"message", err.Error(),
				"value", lengthExtract))
		reply(BadLength)
		return
	}

	capacityExtract := extractPromoInfo(fields, fieldCapacity)
	capacity, err := strToInt(capacityExtract)
	if err != nil {
		log.Error("failed to parse capacity",
			slog.Group("error",
				"message", err.Error(),
				"value", capacityExtract))
		reply(BadCapacity)
		return
	}

	modelToRepo, err := model.NewPromo(promoCode, length, capacity, nil)
	if err != nil {
		log.Error("failed to create promo model",
			slog.Group("error",
				"message", err.Error(),
				"promo_code", promoCode))
		reply(errToCreatePromo)
		return
	}

	switch confirmAct {
	case actionCreate:
		err := h.txManager.WithinTransaction(h.appEnv.Ctx, func(ctx context.Context, q repo.DBQuerier) error {
			ctxTr, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := h.PromoService.CreatePromo(ctxTr, q, modelToRepo); err != nil {
				return fmt.Errorf("create promo: %w", err)
			}

			auditLog := audit.Log{
				Code:   promoCode,
				Action: "create",
				By:     string(opts.UserName),
			}

			if err := h.auditStorage.Save(auditLog); err != nil {
				return fmt.Errorf("audit save: %w", err)
			}

			return nil // Успех -> Commit
		})

		if err != nil {
			log.Error("failed to process promo creation transaction",
				slog.Group("error",
					"message", err.Error(),
					"promo_code", promoCode))
			reply(errToCreatePromo)
			return
		}

		message := fmt.Sprintf(
			reqenv.Lang.Tr(fullMsg),
			promoCode,
			lengthExtract,
			capacityExtract,
		)
		reply(message)

	case actionCancel:
		reply(promoCanceled)
	default:
		reply(UnknowCommand)
	}
}

func extractPromoInfo(fields wizard.Fields, field string) string {
	const op = "extractPromoInfo"
	log := slog.With("op", op, "field", field)

	fieldExtracted := fields.FindField(field)

	if fieldExtracted == nil {
		log.Error("field not found in wizard fields",
			slog.Group("error",
				"message", "nil field"))
		return ""
	}

	var fieldExtractedOut string
	if p, ok := fieldExtracted.Data.(wizard.Txt); ok {
		fieldExtractedOut = p.Value
	} else {
		log.Error("failed to cast field data to wizard.Txt",
			slog.Group("error",
				"message", "type assertion failed",
				"actual_type", fmt.Sprintf("%T", fieldExtracted.Data)))
	}

	return fieldExtractedOut
}

func strToInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.Atoi(s)
}
