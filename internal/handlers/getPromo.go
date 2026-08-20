package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/config"
	"github.com/MikebangSfilya/promoBot/internal/formatter"
	"github.com/MikebangSfilya/promoBot/internal/handlers/common"
	"github.com/MikebangSfilya/promoBot/internal/model"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

const (
	noPromo                   = "noPromo"
	listPromoCodesTitle       = "listPromoCodesTitle"
	listPromoCodesTotalEnding = "listPromoCodesTotalEnding"
	listPromoCodesPageEnding  = "listPromoCodesPageEnding"
	listPromoSortCode         = "listPromoSortCode"
	listPromoSortCapacity     = "listPromoSortCapacity"
	listPromoSortSince        = "listPromoSortSince"
	paginationExpired         = "paginationExpired"
	getCallbackPrefix         = "get:"
	promoPageSize             = 20
)

var errInvalidPromoPage = errors.New("invalid promo page")

type TableGetter interface {
	GetTable(ctx context.Context, limit, offset int, sort model.PromoSort, descending bool, codes ...string) ([]model.ResponseCode, int, error)
}

type promoPageRequest struct {
	page       int
	sort       model.PromoSort
	descending bool
	filtered   bool
}

type GetHandle struct {
	base.CommandHandlerTrait
	common.PrivateCommandTrait

	appEnv *base.ApplicationEnv

	PromoService TableGetter
}

func NewGetHandler(appEnv *base.ApplicationEnv, service TableGetter) *GetHandle {
	h := &GetHandle{
		appEnv:       appEnv,
		PromoService: service,
	}
	h.HandlerRefForTrait = h
	return h
}

func (*GetHandle) GetCommands() []string {
	return []string{"get", "info"}
}

func (h *GetHandle) CallbackHandler() base.CallbackHandler {
	return &getCallbackHandler{get: h}
}

func (h *GetHandle) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	const op = "GetHandle.Handle"
	log := slog.With("op", op, "user_id", msg.From.ID)

	reply := base.NewReplier(h.appEnv, reqEnv, msg)
	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to cast Options to UserOptions",
			slog.Group("error",
				"message", "type assertion failed"))
		reply("failure")
		return
	}

	if opts.Role != config.Admin {
		reply(errNoPermission)
		return
	}
	codes := parseArguments(msg.CommandArguments())
	page := promoPageRequest{
		sort:     model.PromoSortCapacity,
		filtered: len(codes) > 0,
	}
	text, keyboard, err := h.renderPage(reqEnv, page, codes)
	if err != nil {
		log.Error("failed to get promo codes", "error", err)
		reply(failure)
		return
	}
	if len(keyboard.InlineKeyboard) == 0 {
		h.appEnv.Bot.Reply(msg, text)
		return
	}
	h.appEnv.Bot.ReplyWithMessageCustomizer(msg, text, func(config *tgbotapi.MessageConfig) {
		config.ReplyMarkup = keyboard
	})
}

func (h *GetHandle) renderPage(reqEnv *base.RequestEnv, page promoPageRequest, codes []string) (string, tgbotapi.InlineKeyboardMarkup, error) {
	if page.page < 0 || page.page > int(^uint(0)>>1)/promoPageSize || !validPromoSort(page.sort) {
		return "", tgbotapi.InlineKeyboardMarkup{}, errInvalidPromoPage
	}
	offset := page.page * promoPageSize
	ctx, cancel := context.WithTimeout(h.appEnv.Ctx, 10*time.Second)
	defer cancel()

	items, total, err := h.PromoService.GetTable(ctx, promoPageSize, offset, page.sort, page.descending, codes...)
	if err != nil {
		return "", tgbotapi.InlineKeyboardMarkup{}, err
	}
	if total == 0 {
		return reqEnv.Lang.Tr(noPromo), tgbotapi.InlineKeyboardMarkup{}, nil
	}

	pages := (total + promoPageSize - 1) / promoPageSize
	if page.page >= pages {
		return "", tgbotapi.InlineKeyboardMarkup{}, errInvalidPromoPage
	}
	footer := fmt.Sprintf(reqEnv.Lang.Tr(listPromoCodesPageEnding), page.page+1, pages, total)
	list := formatter.FormatPage(
		reqEnv.Lang.Tr(listPromoCodesTitle),
		footer,
		items,
		offset)

	return list, paginationKeyboard(reqEnv, page, pages), nil
}

func paginationKeyboard(reqEnv *base.RequestEnv, page promoPageRequest, pages int) tgbotapi.InlineKeyboardMarkup {
	sortOptions := []struct {
		sort model.PromoSort
		key  string
	}{
		{model.PromoSortCode, listPromoSortCode},
		{model.PromoSortCapacity, listPromoSortCapacity},
		{model.PromoSortSince, listPromoSortSince},
	}

	sortRow := make([]tgbotapi.InlineKeyboardButton, 0, len(sortOptions))
	for _, option := range sortOptions {
		label := reqEnv.Lang.Tr(option.key)
		descending := false
		if page.sort == option.sort {
			if page.descending {
				label += " ↓"
			} else {
				label += " ↑"
			}
			descending = !page.descending
		}
		data := formatGetCallback(promoPageRequest{
			sort:       option.sort,
			descending: descending,
			filtered:   page.filtered,
		})
		sortRow = append(sortRow, tgbotapi.NewInlineKeyboardButtonData(label, data))
	}

	rows := [][]tgbotapi.InlineKeyboardButton{sortRow}
	navigation := make([]tgbotapi.InlineKeyboardButton, 0, 2)
	if page.page > 0 {
		previous := page
		previous.page--
		navigation = append(navigation, tgbotapi.NewInlineKeyboardButtonData("⬅️", formatGetCallback(previous)))
	}
	if page.page+1 < pages {
		next := page
		next.page++
		navigation = append(navigation, tgbotapi.NewInlineKeyboardButtonData("➡️", formatGetCallback(next)))
	}
	if len(navigation) > 0 {
		rows = append(rows, navigation)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func formatGetCallback(page promoPageRequest) string {
	return fmt.Sprintf("%s%d:%s:%t:%t", getCallbackPrefix, page.page, page.sort, page.descending, page.filtered)
}

func parseGetCallback(data string) (promoPageRequest, error) {
	parts := strings.Split(data, ":")
	if len(parts) != 5 || parts[0]+":" != getCallbackPrefix {
		return promoPageRequest{}, errInvalidPromoPage
	}
	page, err := strconv.Atoi(parts[1])
	if err != nil {
		return promoPageRequest{}, errInvalidPromoPage
	}
	sort := model.PromoSort(parts[2])
	if !validPromoSort(sort) {
		return promoPageRequest{}, errInvalidPromoPage
	}
	descending, err := strconv.ParseBool(parts[3])
	if err != nil {
		return promoPageRequest{}, errInvalidPromoPage
	}
	filtered, err := strconv.ParseBool(parts[4])
	if err != nil {
		return promoPageRequest{}, errInvalidPromoPage
	}
	return promoPageRequest{page: page, sort: sort, descending: descending, filtered: filtered}, nil
}

func validPromoSort(sort model.PromoSort) bool {
	return sort == model.PromoSortCode || sort == model.PromoSortCapacity || sort == model.PromoSortSince
}

type getCallbackHandler struct {
	get *GetHandle
}

func (*getCallbackHandler) GetCallbackPrefix() string {
	return getCallbackPrefix
}

func (h *getCallbackHandler) Handle(reqEnv *base.RequestEnv, query *tgbotapi.CallbackQuery) {
	const op = "getCallbackHandler.Handle"
	log := slog.With("op", op, "user_id", query.From.ID)
	answer := func(key string) {
		text := ""
		if key != "" {
			text = reqEnv.Lang.Tr(key)
		}
		if err := h.get.appEnv.Bot.Request(tgbotapi.NewCallback(query.ID, text)); err != nil {
			log.Error("failed to answer callback", "error", err)
		}
	}

	opts, ok := reqEnv.Options.(config.UserOptions)
	if !ok {
		log.Error("failed to cast Options to UserOptions")
		answer(failure)
		return
	}
	if opts.Role != config.Admin {
		answer(errNoPermission)
		return
	}
	if query.Message == nil {
		answer(paginationExpired)
		return
	}

	page, err := parseGetCallback(query.Data)
	if err != nil {
		answer(paginationExpired)
		return
	}
	var codes []string
	if page.filtered {
		if query.Message.ReplyToMessage == nil {
			answer(paginationExpired)
			return
		}
		codes = parseArguments(query.Message.ReplyToMessage.CommandArguments())
		if len(codes) == 0 {
			answer(paginationExpired)
			return
		}
	}

	text, keyboard, err := h.get.renderPage(reqEnv, page, codes)
	if err != nil {
		log.Error("failed to render promo page", "error", err)
		if errors.Is(err, errInvalidPromoPage) {
			answer(paginationExpired)
		} else {
			answer(failure)
		}
		return
	}
	answer("")
	edit := tgbotapi.NewEditMessageTextAndMarkup(query.Message.Chat.ID, query.Message.MessageID, text, keyboard)
	if err := h.get.appEnv.Bot.Request(edit); err != nil {
		log.Error("failed to edit promo page", "error", err)
	}
}
