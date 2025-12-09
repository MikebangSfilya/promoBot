package model

import (
	"errors"
	"strings"
	"time"
)

var (
	errEmpryCode  = errors.New("code is empty")
	errMinusCap   = errors.New("the capacity cannot be less than zero")
	errZeroLenght = errors.New("Bonus lenght cant be zero")
)

// extractPromoInfo(fields wizard.Fields) в ./promo.go
type PromoCode struct {
	Code        string
	BonusLength int
	Since       time.Time
	Until       *time.Time
	Capacity    int
}

func New(code string, bonusLen, capacity int, until *time.Time) (PromoCode, error) {

	trimCode := strings.TrimSpace(code)
	if trimCode == "" {
		return PromoCode{}, errEmpryCode
	}
	if capacity < 0 {
		return PromoCode{}, errMinusCap
	}
	if bonusLen == 0 {
		return PromoCode{}, errZeroLenght
	}

	var untilTime time.Time
	if until == nil {
		untilTime = time.Now().Add(30 * 24 * time.Hour)
	} else {
		untilTime = *until
	}
	return PromoCode{
		Code:        trimCode,
		BonusLength: bonusLen,
		Until:       &untilTime,
		Capacity:    capacity,
	}, nil

}
