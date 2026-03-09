package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
)

// flexDate unmarshals JSON date fields from multiple formats:
//   - full RFC3339 timestamp: "2026-05-01T00:00:00Z"
//   - date-only string: "2026-05-01" or "01.05.2026"
//   - integer days from now as JSON number: 30
//   - integer days from now as JSON string: "30"
type flexDate time.Time

func (f *flexDate) UnmarshalJSON(b []byte) error {
	var days int
	if err := json.Unmarshal(b, &days); err == nil {
		*f = flexDate(*daysFromNow(days))
		return nil
	}

	var dateStr string
	if err := json.Unmarshal(b, &dateStr); err != nil {
		return err
	}

	if t, err := parseDate(dateStr); err == nil {
		*f = flexDate(*t)
		return nil
	}

	return fmt.Errorf("invalid date format: %s", dateStr)
}

func (f *flexDate) toTime() *time.Time {
	if f == nil {
		return nil
	}
	t := time.Time(*f)
	return &t
}

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
		Since       *flexDate `json:"since"`
		Until       *flexDate `json:"until"`
		Code        string    `json:"code"`
		BonusLength int       `json:"bonus_length"`
		Capacity    int       `json:"capacity"`
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

		code, err := model.NewPromo(req.Code, req.BonusLength, req.Capacity, req.Since.toTime(), req.Until.toTime())
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
