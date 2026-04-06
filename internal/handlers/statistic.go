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
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	CodesField  = "code"
	failure     = "failure"
	invalidArgs = "invalid"
)

type StatsGetter interface {
	GetStats(ctx context.Context, codes ...string) ([]model.StatResponseCode, error)
}

type Stats struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	PromoService StatsGetter
}

func NewStats(appEnv *base.ApplicationEnv, stateStorage wizard.StateStorage, service StatsGetter) *Stats {
	h := &Stats{
		appEnv:       appEnv,
		stateStorage: stateStorage,
		PromoService: service,
	}
	h.HandlerRefForTrait = h
	return h
}

func (*Stats) GetCommands() []string {
	return []string{"stats"}
}

func (h *Stats) GetWizardEnv() *wizard.Env {
	return wizard.NewEnv(h.appEnv, h.stateStorage)
}

func (h *Stats) GetWizardDescriptor() *wizard.FormDescriptor {
	desc := wizard.NewWizardDescriptor(h.action)
	desc.AddField(CodesField, promoFieldsTrPrefix+CodesField)
	return desc
}

func (h *Stats) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	const op = "stats.Handle"
	log := slog.With("op", op)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to get User options",
			slog.Group("error",
				slog.String("message", "type assertion failed")))
		reply("failure")
		return
	}

	if opts.Role != config.Admin {
		reply(errNoPermission)
		return
	}

	codesInput := msg.CommandArguments()

	if len(codesInput) > 0 {
		h.processAndReplyPromoList(reqEnv, msg, codesInput, op)
	} else {
		statsForm := wizard.NewWizard(h, 1)
		statsForm.AddEmptyField(CodesField, wizard.Text)
		statsForm.ProcessNextField(reqEnv, msg)
	}
}

func (h *Stats) action(reqEnv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	const op = "stats.action"
	codesInput := fields.FindField(CodesField).Data.(wizard.Txt).Value
	h.processAndReplyPromoList(reqEnv, msg, codesInput, op)
}

func (h *Stats) processAndReplyPromoList(reqEnv *base.RequestEnv, msg *tgbotapi.Message, input string, op string) {
	log := slog.With("op", op, "user_id", msg.From.ID)
	reply := base.NewReplier(h.appEnv, reqEnv, msg)

	argSlice := parseArguments(input)
	if len(argSlice) == 0 {
		reply(invalidArgs)
		return
	}

	ctx, cancel := context.WithTimeout(h.appEnv.Ctx, 10*time.Second)
	defer cancel()

	codes, err := h.PromoService.GetStats(ctx, argSlice...)
	if err != nil {
		log.Error("failed to get promo code",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("field", CodesField),
				slog.String("input", input),
			))
		reply(failure)
		return
	}

	if len(codes) == 0 {
		reply(noPromo)
		return
	}

	list := formatter.FormatList(
		reqEnv.Lang.Tr(listPromoCodesTitle),
		reqEnv.Lang.Tr(listPromoCodesTotalEnding),
		codes)

	reply(list)
}
