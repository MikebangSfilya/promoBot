package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/formatter"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/model"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

const (
	noPromo                   = "noPromo"
	listPromoCodesTitle       = "listPromoCodesTitle"
	listPromoCodesTotalEnding = "listPromoCodesTotalEnding"
)

type TableGetter interface {
	GetTable(ctx context.Context) ([]model.ResponseCode, error)
}

type GetHandle struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv *base.ApplicationEnv

	PromoService TableGetter
}

func NewGetHandler(appEnv *base.ApplicationEnv, service TableGetter) *GetHandle {
	h := &GetHandle{
		appEnv:       appEnv,
		PromoService: service,
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

	ctx, cancel := context.WithTimeout(h.appEnv.Ctx, 10*time.Second)
	defer cancel()

	promoCodes, err := h.PromoService.GetTable(ctx)
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
