package handlers

import (
	"fmt"
	"strings"

	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

var noPromo = "noPromo"

const (
	listPromoCodesTitle       = "listPromoCodesTitle"
	listPromoCodesTotalEnding = "listPromoCodesTotalEnding"
)

type GetHandle struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv *base.ApplicationEnv

	PromoService *repo.Promo
}

func NewGetHandler(appEnv *base.ApplicationEnv) *GetHandle {
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

	sb := strings.Builder{}
	sb.WriteString(reqEnv.Lang.Tr(listPromoCodesTitle))
	sb.WriteString(": \n\n")

	for i, promo := range promoCodes {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, promo.String()))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(
		reqEnv.Lang.Tr(listPromoCodesTotalEnding),
		len(promoCodes),
	))

	reply(sb.String())
}
