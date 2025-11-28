package handlers

import (
	"testing"

	"github.com/kozalosev/goSadTgBot/wizard"
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
			data:     wizard.Txt{Value: "CAPACITY_100"},
			expected: "CAPACITY_100",
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
