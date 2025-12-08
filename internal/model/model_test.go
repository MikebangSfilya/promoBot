package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel(t *testing.T) {

	testCases := []struct {
		name      string
		promocode string
		length    int
		capacity  int
		wantErr   bool
		err       error
	}{
		{
			name:      "valid_model",
			promocode: "promocodeTest",
			length:    50,
			capacity:  10,
		},
		{
			name:      "valid_model_with_spaces",
			promocode: "promocode      Test",
			length:    50,
			capacity:  10,
		},
		{
			name:      "empty_title",
			promocode: "",
			length:    50,
			capacity:  10,
			wantErr:   true,
			err:       errEmpryCode,
		},
		{
			name:      "zero_lenght",
			promocode: "promocodeTest",
			length:    0,
			capacity:  10,
			wantErr:   true,
			err:       errZeroLenght,
		},
		{
			name:      "zero_capacity",
			promocode: "promocodeTest",
			length:    50,
			capacity:  0,
			wantErr:   false,
		},
		{
			name:      "minus_capacity",
			promocode: "promocodeTest",
			length:    0,
			capacity:  -20,
			wantErr:   true,
			err:       errMinusCap,
		},
		{
			name:      "zero_capacity_and_lenght",
			promocode: "promocodeTest",
			length:    0,
			capacity:  0,
			wantErr:   true,
			err:       errZeroLenght,
		},
		{
			name:      "minus_lenght",
			promocode: "promocodeTest",
			length:    -20000,
			capacity:  5,
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.wantErr {

				promo, err := New(tc.promocode, tc.length, tc.capacity, nil)
				var baseTime = time.Now().Add(30 * 24 * time.Hour)

				assert.NoError(t, err)
				require.NotNil(t, promo)
				require.Equal(t, promo.Code, tc.promocode)
				require.Equal(t, promo.BonusLength, tc.length)
				require.Equal(t, promo.Capacity, tc.capacity)
				require.WithinDuration(t, *promo.Until, baseTime, time.Second)
			} else {
				promo, err := New(tc.promocode, tc.length, tc.capacity, nil)
				require.Error(t, err)
				assert.NotNil(t, promo)
				require.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
