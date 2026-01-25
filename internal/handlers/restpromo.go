package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/db/repo"
	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/kozalosev/goSadTgBot/base"
)

type OneTimePromoHandler struct {
	appEnv *base.ApplicationEnv
	promo  *repo.Promo
	audit  audit.Storage
	tx     repo.TxManager
}

func NewOneTimePromoHandler(appEnv *base.ApplicationEnv, repo *repo.Promo, audit audit.Storage, manager repo.TxManager) *OneTimePromoHandler {
	return &OneTimePromoHandler{
		appEnv: appEnv,
		promo:  repo,
		audit:  audit,
		tx:     manager,
	}
}

func (h *OneTimePromoHandler) GeneratePromo() http.HandlerFunc {
	//code, bonus_length, since, until, capacity
	const op = "OneTimePromoHandler.GeneratePromo"
	log := slog.With("op", op)
	type CreateRequest struct {
		Until       *time.Time `json:"until"`
		Code        string     `json:"code"`
		BonusLength int        `json:"bonus_length"`
		Capacity    int        `json:"capacity"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Error("cant decode request body",
				slog.Group("Error",
					slog.String("reason", err.Error()),
					slog.String("time", time.Now().String())))
			return
		}

		code, err := model.NewPromo(req.Code, req.BonusLength, req.Capacity, req.Until)
		if err != nil {
			log.Error("fail to create promo",
				slog.Group("Error",
					slog.String("message", err.Error()),
					slog.String("promo_code", req.Code)))
			http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		ctxTx, cancel := context.WithTimeout(h.appEnv.Ctx, 10*time.Second)
		defer cancel()

		err = h.tx.WithinTransaction(ctxTx, func(ctx context.Context, q repo.DBQuerier) error {
			ctxTxCr, cancel := context.WithTimeout(ctxTx, 5*time.Second)
			defer cancel()

			if err := h.promo.CreatePromo(ctxTxCr, q, code); err != nil {
				log.Error("fail to create promo",
					slog.Group("Error",
						slog.String("message", err.Error()),
						slog.String("promo_code", req.Code),
						slog.String("time", time.Now().String())))
				http.Error(w, "Promo creation failed: "+err.Error(), http.StatusInternalServerError)
				return fmt.Errorf("%s: %w", op, err)
			}

			auditLog := audit.Log{
				Code:   req.Code,
				Action: "create",
				By:     "auto",
			}

			if err := h.audit.Save(auditLog); err != nil {
				log.Error("fail to save audit log",
					slog.Group("Error",
						slog.String("message", err.Error()),
						slog.String("promo_code", req.Code),
						slog.String("time", time.Now().String())))
				http.Error(w, "Save audit log failed: "+err.Error(), http.StatusInternalServerError)
				return fmt.Errorf("%s: %w", op, err)
			}

			return nil
		})
		if err != nil {
			log.Error("fail to create promo",
				slog.Group("Error",
					slog.String("message", err.Error()),
					slog.String("time", time.Now().String())))
			http.Error(w, "Promo creation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
