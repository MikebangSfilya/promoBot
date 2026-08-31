package handlers

import (
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPromoInfo(t *testing.T) {

	testCases := []struct {
		name     string
		field    string
		data     wizard.Txt
		expected string
	}{
		{
			name:     "promo field",
			field:    fieldPromo,
			data:     wizard.Txt{Value: "PROMO123"},
			expected: "PROMO123",
		},
		{
			name:     "confirmation field",
			field:    fieldConfirmation,
			data:     wizard.Txt{Value: "CONFIRM_YES"},
			expected: "CONFIRM_YES",
		},
		{
			name:     "capacity field",
			field:    fieldCapacity,
			data:     wizard.Txt{Value: "100"},
			expected: "100",
		},
		{
			name:     "lenght field",
			field:    fieldLength,
			data:     wizard.Txt{Value: "100"},
			expected: "100",
		},
		{
			name:     "setDates field",
			field:    fieldSetDates,
			data:     wizard.Txt{Value: actionSetBothDates},
			expected: actionSetBothDates,
		},
		{
			name:     "since field",
			field:    fieldSince,
			data:     wizard.Txt{Value: "15.04.2026"},
			expected: "15.04.2026",
		},
		{
			name:     "until field",
			field:    fieldUntil,
			data:     wizard.Txt{Value: "2026-05-15"},
			expected: "2026-05-15",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields := wizard.Fields{
				&wizard.Field{
					Name: tc.field,
					Data: tc.data,
				},
			}
			testProm := extractPromoInfo(fields, tc.field)
			require.Equal(t, tc.expected, testProm)
		})
	}
}

func TestPromoErrorKey(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "empty code", err: model.ErrEmptyCode, want: promoCodeRequired},
		{name: "zero length", err: model.ErrZeroLength, want: promoLengthNonZero},
		{name: "zero capacity", err: model.ErrZeroCapacity, want: promoCapacityPositive},
		{name: "past end date", err: model.ErrPastUntil, want: promoUntilNotPast},
		{name: "end before start", err: model.ErrUntilBeforeSince, want: promoUntilAfterSince},
		{name: "technical error", err: assert.AnError, want: errToCreatePromo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, promoErrorKey(tt.err, errToCreatePromo))
		})
	}
}

func TestParseDate(t *testing.T) {
	t.Run("relative_days", func(t *testing.T) {
		result, err := parseDate("3")
		require.NoError(t, err)
		expected := time.Now().Add(3 * 24 * time.Hour)
		assert.WithinDuration(t, expected, *result, time.Second)
	})

	t.Run("dd_mm_yyyy", func(t *testing.T) {
		result, err := parseDate("15.04.2026")
		require.NoError(t, err)
		expected := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, *result)
	})

	t.Run("yyyy_mm_dd", func(t *testing.T) {
		result, err := parseDate("2026-04-15")
		require.NoError(t, err)
		expected := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, *result)
	})

	t.Run("invalid_format", func(t *testing.T) {
		_, err := parseDate("not-a-date")
		require.Error(t, err)
	})
}
