package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	errEmptyCode        = errors.New("code is empty")
	errMinusCap         = errors.New("the capacity cannot be less than zero")
	errZeroLength       = errors.New("bonus length cant be zero")
	errZeroCap          = errors.New("capacity cant be zero")
	errUntilBeforeSince = errors.New("until date must be after since date")
	errPastUntil        = errors.New("until date must not be in the past")
)

type PromoCode struct {
	Code        string
	BonusLength int
	Since       *time.Time
	Until       *time.Time
	Capacity    int
}

type ResponseCode struct {
	Code        string
	BonusLength int
	Capacity    int
}

func (rc ResponseCode) String() string {
	return fmt.Sprintf("%s — %d см (%d активаций)",
		rc.Code, rc.BonusLength, rc.Capacity)
}

type StatResponseCode struct {
	Code            string
	Activations     int
	InitialCapacity int
	BonusLength     int
	Capacity        int
}

func (rc StatResponseCode) String() string {
	return fmt.Sprintf("Промокод: %s — %d см. Осталось использований: %d, (Изначальное кол-во использований: %d, активаций: %d)",
		rc.Code, rc.BonusLength, rc.Capacity, rc.InitialCapacity, rc.Activations)
}

func NewPromo(code string, bonusLen, capacity int, since, until *time.Time) (PromoCode, error) {
	trimCode := strings.TrimSpace(code)
	if trimCode == "" {
		return PromoCode{}, errEmptyCode
	}
	switch {
	case capacity < 0:
		return PromoCode{}, errMinusCap
	case capacity == 0:
		return PromoCode{}, errZeroCap
	}
	if bonusLen == 0 {
		return PromoCode{}, errZeroLength
	}

	if until != nil && until.Before(time.Now()) {
		return PromoCode{}, errPastUntil
	}
	if since != nil && until != nil && until.Before(*since) {
		return PromoCode{}, errUntilBeforeSince
	}

	return PromoCode{
		Code:        trimCode,
		BonusLength: bonusLen,
		Since:       since,
		Until:       until,
		Capacity:    capacity,
	}, nil
}
