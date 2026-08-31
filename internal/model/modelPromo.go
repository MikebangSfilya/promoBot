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

type PromoDeleteResult string

const (
	PromoDeleteResultDeleted  PromoDeleteResult = "deleted"
	PromoDeleteResultDisabled PromoDeleteResult = "disabled"
)

type ResponseCode struct {
	Code        string
	BonusLength int
	Capacity    int
}

type PromoSort string

const (
	PromoSortCode     PromoSort = "code"
	PromoSortCapacity PromoSort = "capacity"
	PromoSortSince    PromoSort = "since"
)

func (rc ResponseCode) Format(format string) string {
	return fmt.Sprintf(format, rc.Code, rc.BonusLength, rc.Capacity)
}

type StatResponseCode struct {
	Code            string
	Activations     int
	InitialCapacity int
	BonusLength     int
	Capacity        int
}

func (rc StatResponseCode) Format(format string) string {
	return fmt.Sprintf(format, rc.Code, rc.BonusLength, rc.Capacity, rc.InitialCapacity, rc.Activations)
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
