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

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
	fieldLength       = "length"
	fieldCapacity     = "capacity"
	fieldPromoCreated = "fieldPromoCreated "

	actionCreate = "actionCreate"
	actionCancel = "actionCancel"
	textToCreate = "textToCreate"

	promoCanceled    = "promoCanceled"
	errToCreatePromo = "errToCreatePromo"
)

type PromoHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	PromoService *repo.Promo
}

func NewPromoHanlder(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) *PromoHandler {
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

// Наши поддерживаемые команды
func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {

	role := reqEnv.Options.(config.UserOptions).Role
	if role == config.Admin {
		promoForm := wizard.NewWizard(h, 4)

		promoForm.AddEmptyField(fieldPromo, wizard.Text)
		promoForm.AddEmptyField(fieldLength, wizard.Text)
		promoForm.AddEmptyField(fieldCapacity, wizard.Text)
		promoForm.AddEmptyField(fieldConfirmation, wizard.Text)

		promoForm.ProcessNextField(reqEnv, msg)
	} else {
		return
	}
}

func (h *PromoHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {

	reply := base.NewReplier(h.appEnv, reqenv, msg)

	promoCode := extractPromoInfo(fields, fieldPromo)
	confirmAct := extractPromoInfo(fields, fieldConfirmation)

	lengthExtr := extractPromoInfo(fields, fieldLength)
	length, err := strToInt(lengthExtr)

	if err != nil {
		reply("bad request length")
		return
	}

	capasityExtr := extractPromoInfo(fields, fieldCapacity)
	capasity, err := strToInt(capasityExtr)
	if err != nil {
		reply("bad request cap")
		return
	}

	modelToRepo, err := model.NewPromo(promoCode, length, capasity, nil)
	if err != nil {
		reply("failed to create model: " + err.Error())
		return
	}

	switch confirmAct {
	case actionCreate:
		err := h.PromoService.CreatePromo(modelToRepo)
		if err != nil {
			reply(errToCreatePromo)
			return
		}
		reply(fmt.Sprintf("%s: %s, %s: %d, %s: %d",
			fieldPromoCreated, promoCode,
			fieldLength, modelToRepo.BonusLength,
			fieldCapacity, modelToRepo.Capacity))

	case actionCancel:
		reply(promoCanceled)
	default:
		reply(UnknowCommand)
	}

}

func extractPromoInfo(fields wizard.Fields, field string) string {

	fieldExtracted := fields.FindField(field)

	if fieldExtracted == nil {
		slog.Error("One of fields is nil", "error", "nil fields")
		return ""
	}

	var fieldExtractedOut string
	if p, ok := fieldExtracted.Data.(wizard.Txt); ok {

		fieldExtractedOut = p.Value
	}

	return fieldExtractedOut
}

func strToInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.Atoi(s)
}
