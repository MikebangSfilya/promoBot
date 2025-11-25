package models

import (
	"time"
)

type Promo struct {
	Code        string    `db:"code"`
	BonusLength int       `db:"bonus_length"`
	Since       time.Time `db:"since"`
	Until       time.Time `db:"until"`
	Capacity    int       `db:"capacity"`
}
