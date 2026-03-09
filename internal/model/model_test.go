package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel(t *testing.T) {

	testCases := []struct {
		name      string
		promoCode string
		length    int
		capacity  int
		wantErr   bool
		err       error
	}{
		{
			name:      "valid_model",
			promoCode: "promocodeTest",
			length:    50,
			capacity:  10,
		},
		{
			name:      "valid_model_with_spaces",
			promoCode: "promoCode      Test",
			length:    50,
			capacity:  10,
		},
		{
			name:      "empty_title",
			promoCode: "",
			length:    50,
			capacity:  10,
			wantErr:   true,
			err:       errEmptyCode,
		},
		{
			name:      "zero_lenght",
			promoCode: "promocodeTest",
			length:    0,
			capacity:  10,
			wantErr:   true,
			err:       errZeroLength,
		},
		{
			name:      "zero_capacity",
			promoCode: "promocodeTest",
			length:    50,
			capacity:  0,
			wantErr:   true,
			err:       errZeroCap,
		},
		{
			name:      "minus_capacity",
			promoCode: "promocodeTest",
			length:    0,
			capacity:  -20,
			wantErr:   true,
			err:       errMinusCap,
		},
		{
			name:      "zero_capacity_and_lenght",
			promoCode: "promocodeTest",
			length:    0,
			capacity:  0,
			wantErr:   true,
			err:       errZeroCap,
		},
		{
			name:      "minus_lenght",
			promoCode: "promocodeTest",
			length:    -20000,
			capacity:  5,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.wantErr {

				promo, err := NewPromo(tc.promoCode, tc.length, tc.capacity, nil, nil)

				assert.NoError(t, err)
				require.NotNil(t, promo)
				require.Equal(t, promo.Code, tc.promoCode)
				require.Equal(t, promo.BonusLength, tc.length)
				require.Equal(t, promo.Capacity, tc.capacity)
				require.Nil(t, promo.Since)
				require.Nil(t, promo.Until)
			} else {
				promo, err := NewPromo(tc.promoCode, tc.length, tc.capacity, nil, nil)
				require.Error(t, err)
				assert.NotNil(t, promo)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestNewPromoWithDates(t *testing.T) {
	t.Run("explicit_since_and_until", func(t *testing.T) {
		since := time.Now().Add(1 * 24 * time.Hour)
		until := time.Now().Add(30 * 24 * time.Hour)

		promo, err := NewPromo("TEST", 10, 5, &since, &until)
		require.NoError(t, err)
		require.NotNil(t, promo.Since)
		require.NotNil(t, promo.Until)
		require.WithinDuration(t, since, *promo.Since, time.Second)
		require.WithinDuration(t, until, *promo.Until, time.Second)
	})

	t.Run("until_in_the_past", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		_, err := NewPromo("TEST", 10, 5, nil, &past)
		require.ErrorIs(t, err, errPastUntil)
	})

	t.Run("until_before_since", func(t *testing.T) {
		since := time.Now().Add(10 * 24 * time.Hour)
		until := time.Now().Add(5 * 24 * time.Hour)
		_, err := NewPromo("TEST", 10, 5, &since, &until)
		require.ErrorIs(t, err, errUntilBeforeSince)
	})

	t.Run("nil_dates_no_defaults", func(t *testing.T) {
		promo, err := NewPromo("TEST", 10, 5, nil, nil)
		require.NoError(t, err)
		require.Nil(t, promo.Since)
		require.Nil(t, promo.Until)
	})
}
