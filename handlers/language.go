package handlers

import (
	"promo-bot/db/repo"
	"promo-bot/handlers/common"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/settings"
	"github.com/kozalosev/goSadTgBot/wizard"

	log "github.com/sirupsen/logrus"
)

const (
	langFieldsTrPrefix = "commands.language.fields."

	fieldLanguage = "language"

	enCode = "en"
	enFlag = "🇺🇸"
	ruCode = "ru"
	ruFlag = "🇷🇺"
)

var supportedLangCodes = []string{enFlag, enCode, ruFlag, ruCode}

type LanguageHandler struct {
	base.CommandHandlerTrait //реализация message handler для визарда
	common.GroupCommandTrait

	appEnv       *base.ApplicationEnv // ApplicationEnv - контейнер с зависимостями уровня приложения (БД, API бота, контекст)
	stateStorage wizard.StateStorage  // redis

	userService *repo.UserService // репозиторий для юзерой, технически он щас не нужен но буду работать с ним
}

func NewLanguageHandler(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) *LanguageHandler {
	h := &LanguageHandler{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		userService:  repo.NewUserService(appEnv),
	}
	h.HandlerRefForTrait = h

	return h
}

// Создадим визардка для языка
// GetWizardEnv создает окружение для визарда с доступом к зависимостям приложения и хранилищу состояний(редис)
func (h *LanguageHandler) GetWizardEnv() *wizard.Env {
	//h.appEnv контейнер для зависимостей приложения
	//h.stateStorage для редиса
	return wizard.NewEnv(h.appEnv, h.stateStorage)
}

// По идее тут мы заполняем уже наши формочки, создаем поле языка и возвращает ответа
func (h *LanguageHandler) GetWizardDescriptor() *wizard.FormDescriptor {
	decs := wizard.NewWizardDescriptor(h.changeLangAction)                 //пользователь выбирает сменить язык, мы делаем форму
	lang := decs.AddField(fieldLanguage, langFieldsTrPrefix+fieldLanguage) // не особо понимаю что тут
	lang.InlineKeyboardAnswers = []string{enFlag, ruFlag}                  //заполняем флагами, если пользователь кликнел на en флаг то вызовем changeLangAction
	return decs
}

func (*LanguageHandler) GetCommands() []string {
	return []string{"language", "lang"}

}

func (h *LanguageHandler) Handle(reqenv *base.RequestEnv, msg *tgbotapi.Message) {

	langForm := wizard.NewWizard(h, 1) // вызов помошника на хендлер с 1 полем

	langForm.AddEmptyField(fieldLanguage, wizard.Text)

	langForm.ProcessNextField(reqenv, msg)
}

func (h *LanguageHandler) changeLangAction(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	langFlag := fields.FindField(fieldLanguage).Data.(string) //получаем флаг языка
	langCode := langFlagToCode(langFlag)                      //по флагу берем код языка
	reply := base.NewReplier(h.appEnv, reqenv, msg)           // создаем функцию ответчик и отправляем в юзер сервис

	err := h.userService.ChangeLanguage(msg.From.ID, settings.LangCode(langCode)) //заносим в базу, заглушка. мб не зарабоает и придется поднимать
	if err != nil {
		log.WithField(logconst.FieldHandler, "LanguageHandler").
			WithField(logconst.FieldMethod, "changeLangAction").
			WithField(logconst.FieldCalledObject, "UserService").
			WithField(logconst.FieldCalledMethod, "ChangeLanguage").
			Error(err)
		reply(failure)
	} else {
		//у нас должен пройти этот кейс
		reply(success)
	}
}

// оформление
func langFlagToCode(flag string) string {
	switch flag {
	case enFlag:
		return enCode
	case ruFlag:
		return ruCode
	default:
		return flag
	}
}
