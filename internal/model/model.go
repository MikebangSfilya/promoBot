package models

import (
	"errors"
	"strings"
	"time"
)

var (
	errEmpryCode = errors.New("code is empty")
	errZeroCap   = errors.New("the capacity cannot be less than zero")
)

// extractPromoInfo(fields wizard.Fields) в ./promo.go
type PromoCode struct {
	Code        string
	BonusLength int
	Since       time.Time
	Until       *time.Time
	Capacity    int
}

func New(code string, bonuesLen, capacity int, until *time.Time) (PromoCode, error) {

	trimCode := strings.TrimSpace(code)
	if trimCode == "" {
		return PromoCode{}, errEmpryCode
	}
	if capacity < 0 {
		return PromoCode{}, errZeroCap
	}

	var untilTime time.Time
	if until == nil {
		untilTime = time.Now().Add(30 * 24 * time.Hour)
	} else {
		untilTime = *until
	}
	return PromoCode{
		Code:        trimCode,
		BonusLength: bonuesLen,
		Until:       &untilTime,
		Capacity:    capacity,
	}, nil

}
