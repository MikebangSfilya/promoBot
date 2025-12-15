package handlers

import (
	"fmt"

	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

type GetHandle struct {
	base.CommandHandlerTrait
	common.GroupCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	PromoService *repo.Promo
}

func (h *GetHandle) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(h.appEnv, h.stateStorage)
}

func (h *GetHandle) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(h.action)
	return desc
}

func NewGetHanlde(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage) *GetHandle {
	h := &GetHandle{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		PromoService: repo.NewPromo(appEnv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (*GetHandle) GetCommands() []string {
	return []string{"get", "info"}
}

func (h *GetHandle) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	h.action(reqEnv, msg, nil)
}

func (h *GetHandle) action(reqenv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	reply := base.NewReplier(h.appEnv, reqenv, msg)
	promoCodes, err := h.PromoService.GetTable()
	if err != nil {
		reply("fail")
		return
	}

	if len(promoCodes) == 0 {
		reply("Нет промокодов в базе")
		return
	}

	response := "Промокоды: \n\n"
	for i, promo := range promoCodes {
		response += fmt.Sprintf("%d. %s \n", i+1, promo.String())
	}
	response += fmt.Sprintf("\n Всего: %d промокодов", len(promoCodes))

	reply(response)

}
