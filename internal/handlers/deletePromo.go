package handlers

import (
	"context"
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

const (
	promoDeleted     = "promoDeleted"
	errToDeletePromo = "errToDeletePromo"
)

type DeleteService interface {
	DeletePromoWithAudit(ctx context.Context, code string, auditLog audit.Log) error
}

type DeleteHandler struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv        *base.ApplicationEnv
	deleteService DeleteService
}

func NewDeleteHandler(appEnv *base.ApplicationEnv, service DeleteService) *DeleteHandler {
	h := &DeleteHandler{
		appEnv:        appEnv,
		deleteService: service,
	}
	h.HandlerRefForTrait = h
	return h
}

func (*DeleteHandler) GetCommands() []string {
	return []string{"delete", "del", "remove"}
}

func (h *DeleteHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	const op = "DeleteHandler.Handle"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to cast Options to UserOptions",
			slog.Group("error",
				slog.String("message", "type assertion failed")))
		reply("failure")
		return
	}

	if opts.Role != config.Admin {
		reply(errNoPermission)
		return
	}

	args := parseArguments(msg.CommandArguments())
	if len(args) != 1 {
		reply(invalidArgs)
		return
	}

	code := args[0]
	auditLog := audit.Log{
		Code:   code,
		Action: "delete",
		By:     string(opts.UserName),
	}
	if err := h.deleteService.DeletePromoWithAudit(h.appEnv.Ctx, code, auditLog); err != nil {
		log.Error("failed to process promo delete transaction",
			slog.Group("error",
				"message", err.Error(),
				"promo_code", code))
		reply(errToDeletePromo)
		return
	}

	reply(reqEnv.Lang.Tr(promoDeleted))
}
