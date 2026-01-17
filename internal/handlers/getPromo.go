package handlers

import (
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/handlers/formatter"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

const (
	noPromo                   = "noPromo"
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

	sb := formatter.FormatList(
		reqEnv.Lang.Tr(listPromoCodesTitle),
		reqEnv.Lang.Tr(listPromoCodesTotalEnding),
		promoCodes)
	reply(sb)
}
