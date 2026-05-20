package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type formCall struct {
	name  string
	value interface{}
	empty bool
}

type promoFormStub struct {
	calls []formCall
}

func (s *promoFormStub) AddEmptyField(name string, _ wizard.FieldType) {
	s.calls = append(s.calls, formCall{name: name, empty: true})
}

func (s *promoFormStub) AddPrefilledField(name string, value interface{}) {
	s.calls = append(s.calls, formCall{name: name, value: value})
}

func (s *promoFormStub) AddPrefilledAutoField(string, *tgbotapi.Message) {}
func (s *promoFormStub) AllRequiredFieldsFilled() bool                   { return false }
func (s *promoFormStub) ProcessNextField(*base.RequestEnv, *tgbotapi.Message) {
}

func TestAddPromoFieldsFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []formCall
	}{
		{
			name: "no_args",
			args: nil,
			want: []formCall{
				{name: fieldPromo, empty: true},
				{name: fieldLength, empty: true},
				{name: fieldCapacity, empty: true},
				{name: fieldSetDates, empty: true},
				{name: fieldSince, empty: true},
				{name: fieldUntil, empty: true},
			},
		},
		{
			name: "code_length_capacity",
			args: []string{"PROMO", "10", "5"},
			want: []formCall{
				{name: fieldPromo, value: "PROMO"},
				{name: fieldLength, value: "10"},
				{name: fieldCapacity, value: "5"},
				{name: fieldSetDates, empty: true},
				{name: fieldSince, empty: true},
				{name: fieldUntil, empty: true},
			},
		},
		{
			name: "temporary_until",
			args: []string{"PROMO", "10", "5", "30"},
			want: []formCall{
				{name: fieldPromo, value: "PROMO"},
				{name: fieldLength, value: "10"},
				{name: fieldCapacity, value: "5"},
				{name: fieldSetDates, value: actionFromNowToDate},
				{name: fieldSince, empty: true},
				{name: fieldUntil, value: "30"},
			},
		},
		{
			name: "both_dates",
			args: []string{"PROMO", "10", "5", "2026-05-01", "2026-06-01"},
			want: []formCall{
				{name: fieldPromo, value: "PROMO"},
				{name: fieldLength, value: "10"},
				{name: fieldCapacity, value: "5"},
				{name: fieldSetDates, value: actionSetBothDates},
				{name: fieldSince, value: "2026-05-01"},
				{name: fieldUntil, value: "2026-06-01"},
			},
		},
		{
			name: "endless",
			args: []string{"PROMO", "10", "5", "endless"},
			want: []formCall{
				{name: fieldPromo, value: "PROMO"},
				{name: fieldLength, value: "10"},
				{name: fieldCapacity, value: "5"},
				{name: fieldSetDates, value: actionEndlessPromo},
				{name: fieldSince, empty: true},
				{name: fieldUntil, empty: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := &promoFormStub{}

			addPromoFieldsFromArgs(stub, tt.args)

			require.Len(t, stub.calls, len(tt.want))
			assert.Equal(t, tt.want, stub.calls)
		})
	}
}
