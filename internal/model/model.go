package models

import (
	"time"
)

// пока в str, вместо int, потом будет переделано как нужно как будет исправлена функция
// extractPromoInfo(fields wizard.Fields) в ./promo.go
type Promo struct {
	Code        string    `db:"code"`
	BonusLength int       `db:"bonus_length"`
	Since       time.Time `db:"since"`
	Until       time.Time `db:"until"`
	Capacity    int       `db:"capacity"`
}
