package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/config"
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
	BadSince    = "BadSince"
	BadUntil    = "BadUntil"

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
	fieldLength       = "length"
	fieldCapacity     = "capacity"
	fieldSetDates     = "setDates"
	fieldSince        = "since"
	fieldUntil        = "until"
	fullMsg           = "messages.promo.created_full"

	actionCreate        = "actionCreate"
	actionCancel        = "actionCancel"
	actionSetBothDates  = "actionSetBothDates"
	actionEndlessPromo  = "actionEndlessPromo"
	actionFromNowToDate = "actionFromNowToDate"
	textToCreate        = "textToCreate"

	promoCanceled    = "promoCanceled"
	errToCreatePromo = "errToCreatePromo"
	errNoPermission  = "errNoPermission"
)

type SaveService interface {
	CreatePromoWithAudit(ctx context.Context, modelToRepo model.PromoCode, auditLog audit.Log) error
}

type PromoHandler struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	saveService SaveService
}

func NewPromoHandler(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage, service SaveService) *PromoHandler {
	h := &PromoHandler{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		saveService:  service,
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

	setDates := desc.AddField(fieldSetDates, promoFieldsTrPrefix+fieldSetDates)
	setDates.InlineKeyboardAnswers = []string{actionSetBothDates, actionEndlessPromo, actionFromNowToDate}

	sinceField := desc.AddField(fieldSince, promoFieldsTrPrefix+fieldSince)
	sinceField.SkipIf = skipUnlessFieldValue{Name: fieldSetDates, Value: actionSetBothDates}

	untilField := desc.AddField(fieldUntil, promoFieldsTrPrefix+fieldUntil)
	untilField.SkipIf = &wizard.SkipOnFieldValue{Name: fieldSetDates, Value: actionEndlessPromo}

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

	promoForm := wizard.NewWizard(h, 7)

	promoForm.AddEmptyField(fieldPromo, wizard.Text)
	promoForm.AddEmptyField(fieldLength, wizard.Text)
	promoForm.AddEmptyField(fieldCapacity, wizard.Text)
	promoForm.AddEmptyField(fieldSetDates, wizard.Text)
	promoForm.AddEmptyField(fieldSince, wizard.Text)
	promoForm.AddEmptyField(fieldUntil, wizard.Text)
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
		reply(errToCreatePromo)
		return
	}

	switch confirmAct {
	case actionCreate:
		auditLog := audit.Log{
			Code:   promoCode,
			Action: "create",
			By:     string(opts.UserName),
		}
		err := h.saveService.CreatePromoWithAudit(h.appEnv.Ctx, modelToRepo, auditLog)

		if err != nil {
			log.Error("failed to process promo creation transaction",
				slog.Group("error",
					"message", err.Error(),
					"promo_code", promoCode))
			reply(errToCreatePromo)
			return
		}

		sinceStr := time.Now().Format("02.01.2006")
		untilStr := reqenv.Lang.Tr("dateEndless")
		if modelToRepo.Since != nil {
			sinceStr = modelToRepo.Since.Format("02.01.2006")
		}
		if modelToRepo.Until != nil {
			untilStr = modelToRepo.Until.Format("02.01.2006")
		}
		message := fmt.Sprintf(
			reqenv.Lang.Tr(fullMsg),
			promoCode,
			lengthExtract,
			capacityExtract,
			sinceStr,
			untilStr,
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

// skipUnlessFieldValue is a wizard.SkipCondition that skips the field
// unless the referenced field has the specified value.
type skipUnlessFieldValue struct {
	Name  string
	Value string
}

func (s skipUnlessFieldValue) ShouldBeSkipped(form *wizard.Form) bool {
	f := form.Fields.FindField(s.Name)
	if f == nil {
		return false
	}
	txtData, ok := f.Data.(wizard.Txt)
	return !ok || txtData.Value != s.Value
}

func daysFromNow(days int) *time.Time {
	t := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	return &t
}

func parseDate(s string) (*time.Time, error) {
	if days, err := strconv.Atoi(s); err == nil {
		return daysFromNow(days), nil
	}
	for _, layout := range []string{"02.01.2006", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid date format: %s", s)
}
