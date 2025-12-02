package handlers

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

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

	unableToSave = "Невозможно сохранить прокод"
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
	//Заполняем новое заполненное поле с помощью action
	desc := wizard.NewWizardDescriptor(h.action)

	//Добавляем "притвественное поле"
	desc.AddField(fieldPromo, promoFieldsTrPrefix+fieldPromo)

	//Размер пиписи
	//Ндао будет потом сделать через мапку strings, но пока так
	desc.AddField(fieldLenght, "Введите длину писи для создания промокода")
	//активации
	desc.AddField(fieldCapacity, "Введите кол-во активаций промокода")

	//Инлайн добавление поля с кнопками активировать и отменить
	confirm := desc.AddField(fieldConfirmation, textToCreate)
	confirm.InlineKeyboardAnswers = []string{actionCreate, actionCancel}
	return desc
}

// Наши поддерживаемые команды
func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {

	//Создание визарда
	promoForm := wizard.NewWizard(h, 4)

	//Создание пустых полей для данных
	//Леня, если будешь смотреть можеь описать ниже нужно ли тут столько пустых, мб как-то можно лучше, возможн пойму раньше сам :-0
	promoForm.AddEmptyField(fieldPromo, wizard.Text)
	promoForm.AddEmptyField(fieldLenght, wizard.Text)
	promoForm.AddEmptyField(fieldCapacity, wizard.Text)
	promoForm.AddEmptyField(fieldConfirmation, wizard.Text)

	//Переход к новому полю
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

	modelToRepo := models.PromoCode{
		Code:        promoCode,
		BonusLength: capasity,
		Capacity:    lenght,
		Since:       time.Now(),
		Until:       setEndTime(nil),
	}

	//Создание ответчика
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
	//Извлекаем из сообщения наши поля "промкод и активацию"
	fieldExtracted := fields.FindField(field)

	//Похоже бесполезная проверка, он всегда что-то вернет
	if fieldExtracted == nil {
		slog.Error("One of fields is nil", "error", "nil fields")
		return ""
	}

	//Получаеем из нашего чата с помощью Wizard.Txt сообщения, конкретно тут ясное дело текст полученный из ответа
	var fieldExtractedOut string
	if p, ok := fieldExtracted.Data.(wizard.Txt); ok {
		//Если это текст то присваиваем промокоду значение полученное
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

// func to create a defolt time
func setEndTime(t *time.Time) *time.Time {
	if t == nil {
		endTime := time.Now().Add(30 * 24 * time.Hour)
		return &endTime
	}
	return t
}
