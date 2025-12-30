package handlers

import (
	"log/slog"
	"strings"
	"unicode"

	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/handlers/formatter"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

const (
	StatsField  = "stats"
	failure     = "failure"
	invalidArgs = "invalid"
)

type Stats struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage
	PromoHandler *PromoHandler

	PromoService *repo.Promo
}

func NewStats(handler *PromoHandler) *Stats {
	h := &Stats{
		appEnv:       handler.appEnv,
		stateStorage: handler.stateStorage,
		PromoService: handler.PromoService,
		PromoHandler: handler,
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
	desc.AddField(StatsField, promoFieldsTrPrefix+StatsField)
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
		argSlice := ParseArguments(codesInput)
		if len(argSlice) == 0 {
			reply(invalidArgs)
			return
		}
		codes, err := h.PromoService.GetPromoCode(argSlice)
		if err != nil {
			log.Error("failed to get promo code",
				slog.Group("error",
					slog.String("message", err.Error()),
					slog.String("field", StatsField),
					slog.String("input", codesInput),
				))
			reply(failure)
			return
		}
		if len(codes) == 0 {
			reply(noPromo)
			return
		}
		sb := formatter.FormatList(
			reqEnv.Lang.Tr(listPromoCodesTitle),
			reqEnv.Lang.Tr(listPromoCodesTotalEnding),
			codes)

		reply(sb)

	} else {
		statsForm := wizard.NewWizard(h, 1)
		statsForm.AddEmptyField(StatsField, wizard.Text)
		statsForm.ProcessNextField(reqEnv, msg)
	}
}

func (h *Stats) action(reqEnv *base.RequestEnv, msg *tgbotapi.Message, fields wizard.Fields) {
	const op = "stats.action"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)

	codesInput := fields.FindField(StatsField).Data.(wizard.Txt).Value
	if codesInput == "" {
		log.Error("field not found",
			slog.Group("error",
				slog.String("message", "nil field"),
				slog.String("field", StatsField),
			))
		reply(failure)
		return
	}

	argSlice := ParseArguments(codesInput)
	if len(argSlice) == 0 {
		reply(invalidArgs)
		return
	}

	codes, err := h.PromoService.GetPromoCode(argSlice)
	if err != nil {
		log.Error("failed to get promo code",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("field", StatsField),
				slog.String("input", codesInput),
			))
		reply(failure)
		return
	}

	list := formatter.FormatList(
		reqEnv.Lang.Tr(listPromoCodesTitle),
		reqEnv.Lang.Tr(listPromoCodesTotalEnding),
		codes)

	reply(list)
}

func ParseArguments(arg string) []string {
	return strings.FieldsFunc(arg, func(r rune) bool {
		return unicode.IsSpace(r) || r == ',' || r == ';'
	})
}
