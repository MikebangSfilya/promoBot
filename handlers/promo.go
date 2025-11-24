package handlers

import (
	"github.com/MikebangSfilya/promoBot/db/repo"
	"github.com/MikebangSfilya/promoBot/handlers/common"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	promoFieldsTrPrefix = "commands.promo.fields."

	fieldPromo        = "promo"
	fieldConfirmation = "confirmation"
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

	confirm := desc.AddField(fieldConfirmation, "")
	confirm.InlineKeyboardAnswers = []string{"", ""}
	return desc
}

func (*PromoHandler) GetCommands() []string {
	return []string{"promo", "code", "generate"}
}

func (h *PromoHandler) Handle(reqEnd *base.RequestEnv, msg *tgbotapi.Message) {
	promoForm := wizard.NewWizard(h, 2)

	promoForm.ProcessNextField(reqEnd, msg)

}

func (h *PromoHandler) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	//extract
	// promoField := fields.FindField(fieldPromo)
	// confirmField := fields.FindField(fieldConfirmation)

	// promoField
	// // confirmData, ok2 := confirmField.Data.(wizard.)

	// reply := base.NewReplier(h.appEnv, reqenv, msg)
	// reply(promoFieldsTrPrefix + fieldPromo)
}
