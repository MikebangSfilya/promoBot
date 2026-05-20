package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/model"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

type UpdateService interface {
	UpdatePromoWithAudit(ctx context.Context, modelToRepo model.PromoCode, auditLog audit.Log) error
}

type EditHandler struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv        *base.ApplicationEnv
	stateStorage  wizard.StateStorage
	updateService UpdateService
}

func NewEditHandler(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage, service UpdateService) *EditHandler {
	h := &EditHandler{
		appEnv:        appEnv,
		stateStorage:  stateStorage,
		updateService: service,
	}
	h.HandlerRefForTrait = h
	return h
}

func (h *EditHandler) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(h.appEnv, h.stateStorage)
}

func (h *EditHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(h.action)
	desc.AddField(fieldPromo, promoFieldsTrPrefix+fieldPromo)
	desc.AddField(fieldLength, promoFieldsTrPrefix+fieldLength)
	desc.AddField(fieldCapacity, promoFieldsTrPrefix+fieldCapacity)

	setDates := desc.AddField(fieldSetDates, promoFieldsTrPrefix+fieldSetDates)
	setDates.InlineKeyboardAnswers = []string{actionSetBothDates, actionEndlessPromo, actionFromNowToDate}

	sinceField := desc.AddField(fieldSince, promoFieldsTrPrefix+fieldSince)
	sinceField.SkipIf = skipUnlessFieldValue{Name: fieldSetDates, Value: actionSetBothDates}

	untilField := desc.AddField(fieldUntil, promoFieldsTrPrefix+fieldUntil)
	untilField.SkipIf = &wizard.SkipOnFieldValue{Name: fieldSetDates, Value: actionEndlessPromo}

	confirm := desc.AddField(fieldConfirmation, textToUpdate)
	confirm.InlineKeyboardAnswers = []string{actionUpdate, actionCancel}
	return desc
}

func (*EditHandler) GetCommands() []string {
	return []string{"edit", "update", "modify"}
}

func (h *EditHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	const op = "EditHandler.Handle"
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

	args := parseArguments(msg.CommandArguments())

	editForm := wizard.NewWizard(h, 7)
	addPromoFieldsFromArgs(editForm, args)
	editForm.AddEmptyField(fieldConfirmation, wizard.Text)
	editForm.ProcessNextField(reqEnv, msg)
}

func (h *EditHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	const op = "EditHandler.action"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqenv, msg)

	opts, ok := reqenv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to get user options",
			slog.Group("error",
				"message", "type assertion failed"))
		reply(errToUpdatePromo)
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

	var (
		since *time.Time
		until *time.Time
	)
	switch extractPromoInfo(fields, fieldSetDates) {
	case actionSetBothDates:
		sinceStr := extractPromoInfo(fields, fieldSince)
		since, err = parseDate(sinceStr)
		if err != nil {
			log.Error("failed to parse since date",
				slog.Group("error",
					"message", err.Error(),
					"value", sinceStr))
			reply(BadSince)
			return
		}

		untilStr := extractPromoInfo(fields, fieldUntil)
		until, err = parseDate(untilStr)
		if err != nil {
			log.Error("failed to parse until date",
				slog.Group("error",
					"message", err.Error(),
					"value", untilStr))
			reply(BadUntil)
			return
		}
	case actionFromNowToDate:
		untilStr := extractPromoInfo(fields, fieldUntil)
		until, err = parseDate(untilStr)
		if err != nil {
			log.Error("failed to parse until date",
				slog.Group("error",
					"message", err.Error(),
					"value", untilStr))
			reply(BadUntil)
			return
		}
	}

	modelToRepo, err := model.NewPromo(promoCode, length, capacity, since, until)
	if err != nil {
		log.Error("failed to create promo model",
			slog.Group("error",
				"message", err.Error(),
				"promo_code", promoCode))
		reply(errToUpdatePromo)
		return
	}

	switch confirmAct {
	case actionUpdate:
		auditLog := audit.Log{
			Code:   promoCode,
			Action: "update",
			By:     string(opts.UserName),
		}
		if err := h.updateService.UpdatePromoWithAudit(h.appEnv.Ctx, modelToRepo, auditLog); err != nil {
			log.Error("failed to process promo update transaction",
				slog.Group("error",
					"message", err.Error(),
					"promo_code", promoCode))
			reply(errToUpdatePromo)
			return
		}
		reply(fmt.Sprintf(reqenv.Lang.Tr(promoUpdated), promoCode))
	case actionCancel:
		reply(promoCanceled)
	default:
		reply(UnknowCommand)
	}
}
