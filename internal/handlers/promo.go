package handlers

import (
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	promoFieldsTrPrefix = "commands.promo.fields."
	// promoCreatePrefix   = "promo.create"

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
	actionCreate      = "Создать"
	actionCancel      = "Отменить"
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

	//Инлайн добавление поля с кнопками активировать и отменить
	confirm := desc.AddField(fieldConfirmation, "Подтвердите создание промокода:")
	confirm.InlineKeyboardAnswers = []string{actionCreate, actionCancel}
	return desc
}

// Наши поддерживаемые команды
func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {

	//Создаем помошника с 2 полями
	promoForm := wizard.NewWizard(h, 2)

	//Создание пустых полей для данных
	promoForm.AddEmptyField(fieldPromo, wizard.Text)
	promoForm.AddEmptyField(fieldConfirmation, wizard.Text)

	//Переход к новому полю
	promoForm.ProcessNextField(reqEnv, msg)

}

func (h *PromoHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	//extract
	//Получаем наши строки промокода и подтвержения
	promoCode, confirmAct := extractPromoInfo(fields)

	//Создание ответчика
	reply := base.NewReplier(h.appEnv, reqenv, msg)

	//Пустой промокод
	if promoCode == "" {
		reply("Ошибка: промокод не может быть пустым")
		return
	}
	switch confirmAct {
	case actionCreate:
		success, err := h.userService.CreatePromo(promoCode)
		if err != nil {
			reply("Ошибка при создание промокода")
			return
		}
		if success {
			reply("Промокод '" + promoCode + "' создан")
		} else {
			reply("Невозможно сохранить прокод")
		}
	case actionCancel:
		reply("Создание промокода отменено")
	default:
		reply("Неизвестное действие")
	}

}

func extractPromoInfo(fields wizard.Fields) (string, string) {
	//Извлекаем из сообщения наши поля "промкод и активацию"
	promoField := fields.FindField(fieldPromo)
	confirmField := fields.FindField(fieldConfirmation)

	if promoField == nil || confirmField == nil {
		slog.Error("One of fields is nil", "error", "nil fields")
		return "", ""
	}

	//Получаеем из нашего чата с помощью Wizard.Txt сообщени
	var promoCode string
	if p, ok := promoField.Data.(wizard.Txt); ok {
		//Если это текст то присваиваем промокоду значение полученное
		promoCode = p.Value
	}

	var action string
	//Получаеем из нашего чата с помощью Wizard.Txt сообщени
	if c, ok := confirmField.Data.(wizard.Txt); ok {
		//Если это текст то присваиваем действию значение полученное
		action = c.Value
	}

	return promoCode, action
}
