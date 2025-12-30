package handlers

import (
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
)

type Stats struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv       *base.ApplicationEnv
	stateStorage wizard.StateStorage

	PromoService *repo.Promo
	PromoHandler
}

func New(handler PromoHandler) *Stats {
	h := &Stats{
		PromoHandler: handler,
	}
	h.HandlerRefForTrait = h
	return h
}
