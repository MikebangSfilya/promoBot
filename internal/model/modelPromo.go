package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmptyCode        = errors.New("code is empty")
	ErrMinusCapacity    = errors.New("the capacity cannot be less than zero")
	ErrZeroLength       = errors.New("bonus length cant be zero")
	ErrZeroCapacity     = errors.New("capacity cant be zero")
	ErrUntilBeforeSince = errors.New("until date must be after since date")
	ErrPastUntil        = errors.New("until date must not be in the past")
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
		return PromoCode{}, ErrEmptyCode
	}
	switch {
	case capacity < 0:
		return PromoCode{}, ErrMinusCapacity
	case capacity == 0:
		return PromoCode{}, ErrZeroCapacity
	}
	if bonusLen == 0 {
		return PromoCode{}, ErrZeroLength
	}

	if until != nil && until.Before(time.Now()) {
		return PromoCode{}, ErrPastUntil
	}
	if since != nil && until != nil && until.Before(*since) {
		return PromoCode{}, ErrUntilBeforeSince
	}

	return PromoCode{
		Code:        trimCode,
		BonusLength: bonusLen,
		Since:       since,
		Until:       until,
		Capacity:    capacity,
	}, nil
}
