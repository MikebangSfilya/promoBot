package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
)

type OneTimePromoHandler struct {
	SaveService SaveService
}

func NewOneTimePromoHandler(service SaveService) *OneTimePromoHandler {
	return &OneTimePromoHandler{
		SaveService: service,
	}
}

func (h *OneTimePromoHandler) GeneratePromo() http.HandlerFunc {
	const op = "OneTimePromoHandler.GeneratePromo"
	log := slog.With("op", op)
	type CreateRequest struct {
		Until       *time.Time `json:"until"`
		Code        string     `json:"code"`
		BonusLength int        `json:"bonus_length"`
		Capacity    int        `json:"capacity"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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
		ctxTx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		auditLog := audit.Log{
			Code:   req.Code,
			Action: "create",
			By:     "auto",
		}
		err = h.SaveService.CreatePromoWithAudit(ctxTx, code, auditLog)
		if err != nil {
			log.Error("fail to create promo",
				slog.Group("Error",
					slog.String("message", err.Error()),
					slog.String("time", time.Now().String())))
			http.Error(w, "Promo creation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "code": code.Code})
	}
}
