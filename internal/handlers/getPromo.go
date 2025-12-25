package handlers

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/MikebangSfilya/promoBot/internal/config"
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
	const op = "GetHandle.Handle"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to cast Options to UserOptions",
			slog.Group("error",
				"message", "type assertion failed"))
		reply("failure")
		return
	}

	if opts.Role != config.Admin {
		reply(errNoPermission)
		return
	}

	promoCodes, err := h.PromoService.GetTable()
	if err != nil {
		log.Error("failed to get promo codes table",
			slog.Group("error",
				"message", err.Error(),
				"component", "PromoService.GetTable"))
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
