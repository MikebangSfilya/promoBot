package handlers

import (
	"context"
	"strings"
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chatScopeLang() *loc.Context {
	pool := loc.NewPool("en")
	pool.Resources["en"] = map[string]string{
		groupCommandsDisabled: "DM only",
	}
	return pool.GetContext("en")
}

func command(chatType, cmd string) *tgbotapi.Message {
	return &tgbotapi.Message{
		MessageID: 1,
		Chat:      tgbotapi.Chat{ID: 20, Type: chatType},
		Text:      "/" + cmd,
		Entities:  []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(cmd) + 1}},
	}
}

func sentTexts(t *testing.T, bot *base.FakeBotAPI) []string {
	t.Helper()
	if bot.GetOutput() == nil {
		return nil
	}
	texts, ok := bot.GetOutput().([]string)
	require.True(t, ok)
	return texts
}

// Commands other than /help must be inert outside direct messages.
func TestPrivateCommandsAreRejectedInGroups(t *testing.T) {
	appEnv := &base.ApplicationEnv{Bot: &base.FakeBotAPI{}, Ctx: context.Background()}
	reqEnv := base.NewRequestEnv(chatScopeLang(), nil)

	handlers := map[string]base.MessageHandler{
		"get":    NewGetHandler(appEnv, &tableGetterStub{}),
		"delete": NewDeleteHandler(appEnv, nil),
	}

	for cmd, handler := range handlers {
		t.Run(cmd, func(t *testing.T) {
			assert.True(t, handler.CanHandle(reqEnv, command("private", cmd)))
			assert.False(t, handler.CanHandle(reqEnv, command("group", cmd)))
			assert.False(t, handler.CanHandle(reqEnv, command("supergroup", cmd)))
		})
	}
}

func TestHelpHandlerAnswersInEveryChatType(t *testing.T) {
	tests := []struct {
		chatType string
		want     string
	}{
		{"private", strings.TrimSpace(helpPrivateEn)},
		{"group", strings.TrimSpace(helpGroupEn)},
		{"supergroup", strings.TrimSpace(helpGroupEn)},
	}

	for _, tt := range tests {
		t.Run(tt.chatType, func(t *testing.T) {
			bot := &base.FakeBotAPI{}
			handler := NewHelpHandler(&base.ApplicationEnv{Bot: bot, Ctx: context.Background()})
			msg := command(tt.chatType, "help")
			reqEnv := base.NewRequestEnv(chatScopeLang(), nil)

			require.True(t, handler.CanHandle(reqEnv, msg))
			handler.Handle(reqEnv, msg)

			assert.Equal(t, []string{tt.want}, sentTexts(t, bot))
		})
	}
}

func TestHelpTextsFor(t *testing.T) {
	assert.Equal(t, helpPrivateEn, helpTextsFor("en").private)
	assert.Equal(t, helpPrivateRu, helpTextsFor("ru").private)
	assert.Equal(t, helpGroupRu, helpTextsFor("de").group, "unknown languages fall back to the pool default")

	for lang, texts := range helpTextsByLang {
		assert.NotEmpty(t, strings.TrimSpace(texts.private), lang)
		assert.NotEmpty(t, strings.TrimSpace(texts.group), lang)
	}
}

func TestHelpHandlerAlsoAnswersStart(t *testing.T) {
	handler := NewHelpHandler(&base.ApplicationEnv{Bot: &base.FakeBotAPI{}, Ctx: context.Background()})

	assert.True(t, handler.CanHandle(nil, command("private", "start")))
	assert.Equal(t, []base.CommandScope{base.CommandScopeDefault}, handler.GetScopes())
}

func TestGroupGuardHandler(t *testing.T) {
	plain := func(chatType string) *tgbotapi.Message {
		return &tgbotapi.Message{MessageID: 1, Chat: tgbotapi.Chat{ID: 20, Type: chatType}, Text: "hello"}
	}

	tests := []struct {
		name       string
		msg        *tgbotapi.Message
		canHandle  bool
		wantTexts  []string
		wantSilent bool
	}{
		{name: "group command is refused", msg: command("group", "get"), canHandle: true, wantTexts: []string{"DM only"}},
		{name: "supergroup command is refused", msg: command("supergroup", "promo"), canHandle: true, wantTexts: []string{"DM only"}},
		{name: "group chatter is ignored", msg: plain("group"), canHandle: true, wantSilent: true},
		{name: "private command is not guarded", msg: command("private", "get"), canHandle: false, wantSilent: true},
		{name: "private message is not guarded", msg: plain("private"), canHandle: false, wantSilent: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := &base.FakeBotAPI{}
			handler := NewGroupGuardHandler(&base.ApplicationEnv{Bot: bot, Ctx: context.Background()})
			reqEnv := base.NewRequestEnv(chatScopeLang(), nil)

			assert.Equal(t, tt.canHandle, handler.CanHandle(reqEnv, tt.msg))
			if !tt.canHandle {
				return
			}

			handler.Handle(reqEnv, tt.msg)
			if tt.wantSilent {
				assert.Empty(t, sentTexts(t, bot))
				return
			}
			assert.Equal(t, tt.wantTexts, sentTexts(t, bot))
		})
	}
}

// The guard is not a CommandHandler, so it never reaches the command menu.
func TestGroupGuardHandlerIsNotACommand(t *testing.T) {
	var handler base.MessageHandler = NewGroupGuardHandler(nil)

	_, isCommand := handler.(base.CommandHandler)
	assert.False(t, isCommand)
}
