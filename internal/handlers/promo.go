package handlers

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	models "github.com/MikebangSfilya/promoBot/internal/model"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	promoFieldsTrPrefix = "commands.promo.fields."
	// promoCreatePrefix   = "promo.create"

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
	fieldLenght       = "lenght"
	fieldCapacity     = "capacity"
	fieldPromoCreated = "Промокод создан: "

	actionCreate = "Создать"
	actionCancel = "Отменить"
	textToCreate = "Подтвердите создание промокода:"

	promoCanceled    = "Создание промокода отменено"
	errUnknowComma   = "Неизвестная команда"
	errToCreatePromo = "Ошибка при создание промокода"

	success = "success"
	failure = "failure"
)

type PromoHandler struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	userService *repo.UserService //TODO переделать в ирл реализацию, заглушка
}

func NewPromoHanlder(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) *PromoHandler {
	h := &PromoHandler{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appEnv),
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

	desc.AddField(fieldLenght, "Введите длину писи для создания промокода")

	desc.AddField(fieldCapacity, "Введите кол-во активаций промокода")

	confirm := desc.AddField(fieldConfirmation, textToCreate)
	confirm.InlineKeyboardAnswers = []string{actionCreate, actionCancel}
	return desc
}

// Наши поддерживаемые команды
func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {

	promoForm := wizard.NewWizard(h, 4)

	promoForm.AddEmptyField(fieldPromo, wizard.Text)
	promoForm.AddEmptyField(fieldLenght, wizard.Text)
	promoForm.AddEmptyField(fieldCapacity, wizard.Text)
	promoForm.AddEmptyField(fieldConfirmation, wizard.Text)

	promoForm.ProcessNextField(reqEnv, msg)

}

func (h *PromoHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	//extract
	promoCode := extractPromoInfo(fields, fieldPromo)
	confirmAct := extractPromoInfo(fields, fieldConfirmation)
	capasityExtr := extractPromoInfo(fields, fieldCapacity)
	lenghtExtr := extractPromoInfo(fields, fieldLenght)

	capasity, err := strToInt(capasityExtr)
	if err != nil {
		return
	}
	lenght, err := strToInt(lenghtExtr)
	if err != nil {
		return
	}

	modelToRepo, err := models.New(promoCode, lenght, capasity, nil)
	if err != nil {
		fmt.Printf("failed to create model %v", err)
		return
	}

	reply := base.NewReplier(h.appEnv, reqenv, msg)

	switch confirmAct {
	case actionCreate:
		err := h.userService.CreatePromo(modelToRepo)
		if err != nil {
			slog.Error(errToCreatePromo, "error", err)
			reply(errToCreatePromo)
			return
		}
		reply(fmt.Sprintf("%s: %s, %s: %d, %s: %d",
			fieldPromoCreated, promoCode,
			fieldLenght, modelToRepo.BonusLength,
			fieldCapacity, modelToRepo.Capacity))

	case actionCancel:
		reply(promoCanceled)
	default:
		reply(errUnknowComma)
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
