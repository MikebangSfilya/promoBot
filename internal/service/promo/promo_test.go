package promo

import (
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/audit"
	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestPromoChanges(t *testing.T) {
	oldSince := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	newSince := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
	oldUntil := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	newUntil := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)

	changes := promoChanges(
		model.PromoCode{
			Code:        "PROMO",
			BonusLength: 10,
			Since:       &oldSince,
			Until:       &oldUntil,
			Capacity:    5,
		},
		model.PromoCode{
			Code:        "PROMO",
			BonusLength: 15,
			Since:       &newSince,
			Until:       &newUntil,
			Capacity:    3,
		},
	)

	assert.Equal(t, map[string]audit.Change{
		"bonus_length": {Old: "10", New: "15"},
		"since":        {Old: "2026-06-01", New: "2026-06-02"},
		"until":        {Old: "2026-07-01", New: "2026-07-05"},
		"capacity":     {Old: "5", New: "3"},
	}, changes)
}

func TestPromoChangesNoDiff(t *testing.T) {
	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	changes := promoChanges(
		model.PromoCode{
			Code:        "PROMO",
			BonusLength: 10,
			Since:       &since,
			Capacity:    5,
		},
		model.PromoCode{
			Code:        "PROMO",
			BonusLength: 10,
			Since:       &since,
			Capacity:    5,
		},
	)

	assert.Nil(t, changes)
}
