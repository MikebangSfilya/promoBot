package handlers

import (
	_ "embed"
	"strings"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

// helpFallbackLang matches the default language of the localization pool.
const helpFallbackLang = "ru"

type helpTexts struct {
	private, group string
}

var (
	//go:embed help/private.md
	helpPrivateEn string
	//go:embed help/private.ru.md
	helpPrivateRu string

	//go:embed help/group.md
	helpGroupEn string
	//go:embed help/group.ru.md
	helpGroupRu string

	// Long multi-paragraph texts live in Markdown files instead of strings.go
	// so they are easy to read and edit.
	helpTextsByLang = map[string]helpTexts{
		"en": {private: helpPrivateEn, group: helpGroupEn},
		"ru": {private: helpPrivateRu, group: helpGroupRu},
	}
)

// HelpHandler explains what the bot does. It is the only command that stays
// available in group chats, so it embeds [base.CommandHandlerTrait] directly
// instead of the private-only trait the other commands use.
type HelpHandler struct {
	base.CommandHandlerTrait

	appEnv *base.ApplicationEnv
}

func NewHelpHandler(appEnv *base.ApplicationEnv) *HelpHandler {
	h := &HelpHandler{appEnv: appEnv}
	h.HandlerRefForTrait = h
	return h
}

func (*HelpHandler) GetCommands() []string {
	return []string{"help", "start"}
}

func (*HelpHandler) GetScopes() []base.CommandScope {
	return []base.CommandScope{base.CommandScopeDefault}
}

// Handle replies with the help text for the chat the command came from.
func (h *HelpHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	texts := helpTextsFor(reqEnv.Lang.GetLanguage())

	text := texts.group
	if msg.Chat.IsPrivate() {
		text = texts.private
	}
	h.appEnv.Bot.Reply(msg, strings.TrimSpace(text))
}

func helpTextsFor(lang string) helpTexts {
	if texts, ok := helpTextsByLang[lang]; ok {
		return texts
	}
	return helpTextsByLang[helpFallbackLang]
}
