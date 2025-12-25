package handlers

import (
	"fmt"
	"log/slog"
	"strconv"

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

	PromoService *repo.Promo
}

func NewPromoHandler(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) *PromoHandler {
	h := &PromoHandler{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		PromoService: repo.NewPromo(appEnv),
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
		log.Error("failed to create a promo code",
			slog.Group("error",
				"message", "type assertion failed"))
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
		reply("failed to create model: " + err.Error())
		return
	}

	switch confirmAct {
	case actionCreate:
		err := h.PromoService.CreatePromo(modelToRepo)
		if err != nil {
			log.Error("failed to create promo code",
				slog.Group("error",
					"message", err.Error(),
					"component", "PromoService.CreatePromo",
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
