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

// Audit actions recorded for promo codes.
//
// The audit storage matches actions verbatim and knows nothing about them; the
// vocabulary belongs to the domain, and these constants keep the handlers that
// write them and the report that reads them from drifting apart.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

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

// PromoActivationStat is what the database knows about a promo code:
// ActivationsInWindow counts only the activations that happened within the
// reported period, while ActivationsTotal counts them since the code was created.
type PromoActivationStat struct {
	Code                string
	BonusLength         int
	Capacity            int
	Since               *time.Time
	Until               *time.Time
	ActivationsInWindow int
	ActivationsTotal    int
}

// ReportEntry is a single promo code as it appears in the weekly activation
// report. The database has no record of who created a code or when, so that
// half comes from the audit log.
type ReportEntry struct {
	PromoActivationStat

	CreatedBy string
	CreatedAt time.Time
}

// InitialCapacity restores the number of activations the code was created with.
// Every activation decrements the remaining capacity, so the sum of the two is
// the original value.
func (s PromoActivationStat) InitialCapacity() int {
	return s.Capacity + s.ActivationsTotal
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
