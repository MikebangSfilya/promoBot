package handlers

import (
	"context"
	"testing"

	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/model"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tableGetterStub struct {
	limit, offset int
	sort          model.PromoSort
	descending    bool
	codes         []string
}

func (s *tableGetterStub) GetTable(_ context.Context, limit, offset int, sort model.PromoSort, descending bool, codes ...string) ([]model.ResponseCode, int, error) {
	s.limit, s.offset, s.sort, s.descending, s.codes = limit, offset, sort, descending, codes
	return []model.ResponseCode{{Code: "RESULT", BonusLength: 1, Capacity: 2}}, 21, nil
}

func TestGetCallbackRoundTrip(t *testing.T) {
	want := promoPageRequest{page: 2, sort: model.PromoSortSince, descending: true, filtered: true}

	got, err := parseGetCallback(formatGetCallback(want))

	require.NoError(t, err)
	assert.Equal(t, want, got)
	_, err = parseGetCallback("get:0:drop table:false:false")
	assert.Error(t, err)
}

func TestPaginationKeyboard(t *testing.T) {
	pool := loc.NewPool("en")
	pool.Resources["en"] = map[string]string{
		listPromoSortCode:     "Code",
		listPromoSortCapacity: "Remaining",
		listPromoSortSince:    "Start date",
	}
	reqEnv := base.NewRequestEnv(pool.GetContext("en"), nil)
	page := promoPageRequest{sort: model.PromoSortCapacity, filtered: true}

	keyboard := paginationKeyboard(reqEnv, page, 3)

	require.Len(t, keyboard.InlineKeyboard, 2)
	require.Len(t, keyboard.InlineKeyboard[0], 3)
	assert.Equal(t, "Remaining ↑", keyboard.InlineKeyboard[0][1].Text)

	toggled, err := parseGetCallback(*keyboard.InlineKeyboard[0][1].CallbackData)
	require.NoError(t, err)
	assert.Equal(t, promoPageRequest{sort: model.PromoSortCapacity, descending: true, filtered: true}, toggled)

	require.Len(t, keyboard.InlineKeyboard[1], 1)
	next, err := parseGetCallback(*keyboard.InlineKeyboard[1][0].CallbackData)
	require.NoError(t, err)
	assert.Equal(t, promoPageRequest{page: 1, sort: model.PromoSortCapacity, filtered: true}, next)
}

func TestGetCallbackRestoresFiltersAndEditsPage(t *testing.T) {
	pool := loc.NewPool("en")
	pool.Resources["en"] = map[string]string{
		listPromoCodesTitle:      "Promo codes",
		listPromoCodesPageEnding: "Page %d/%d · Total: %d",
		listPromoSortCode:        "Code",
		listPromoSortCapacity:    "Remaining",
		listPromoSortSince:       "Start date",
	}
	bot := &base.FakeBotAPI{}
	service := &tableGetterStub{}
	handler := NewGetHandler(&base.ApplicationEnv{Bot: bot, Ctx: context.Background()}, service).CallbackHandler()
	reqEnv := base.NewRequestEnv(pool.GetContext("en"), config.UserOptions{Role: config.Admin})
	query := &tgbotapi.CallbackQuery{
		ID:   "callback",
		From: &tgbotapi.User{ID: 1},
		Data: formatGetCallback(promoPageRequest{page: 1, sort: model.PromoSortCode, descending: true, filtered: true}),
		Message: &tgbotapi.Message{
			MessageID: 10,
			Chat:      tgbotapi.Chat{ID: 20},
			ReplyToMessage: &tgbotapi.Message{
				Text:     "/get foo bar",
				Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: 4}},
			},
		},
	}

	handler.Handle(reqEnv, query)

	assert.Equal(t, promoPageSize, service.limit)
	assert.Equal(t, promoPageSize, service.offset)
	assert.Equal(t, model.PromoSortCode, service.sort)
	assert.True(t, service.descending)
	assert.Equal(t, []string{"foo", "bar"}, service.codes)
	requests := bot.GetOutput().([]tgbotapi.Chattable)
	require.Len(t, requests, 2)
	edit, ok := requests[1].(tgbotapi.EditMessageTextConfig)
	require.True(t, ok)
	assert.Contains(t, edit.Text, "Page 2/2 · Total: 21")
}
