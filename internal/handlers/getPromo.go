package handlers

import (
	"fmt"

	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

var noPromo = "noPromo"

type GetHandle struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv *base.ApplicationEnv

	PromoService *repo.Promo
}

func NewGetHandle(appEnv *base.ApplicationEnv) *GetHandle {
	h := &GetHandle{
		appEnv:       appEnv,
		PromoService: repo.NewPromo(appEnv),
	}
	h.HandlerRefForTrait = h
	return h
}

func (*GetHandle) GetCommands() []string {
	return []string{"get", "info"}
}

func (h *GetHandle) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	promoCodes, err := h.PromoService.GetTable()
	if err != nil {
		reply("failure")
		return
	}

	if len(promoCodes) == 0 {
		reply(noPromo)
		return
	}

	//TODO: переделать в нормальный вид
	response := "Промокоды: \n\n"
	for i, promo := range promoCodes {
		response += fmt.Sprintf("%d. %s \n", i+1, promo.String())
	}
	response += fmt.Sprintf("\nВсего: %d промокодов", len(promoCodes))

	reply(response)
}
