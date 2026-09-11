package common

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

// PrivateCommandTrait restricts a command to private chats.
//
// GetScopes only controls which commands Telegram advertises in its menus, so
// the trait also overrides CanHandle to actually reject the command when it is
// typed in a group. It embeds [base.CommandHandlerTrait] rather than sitting
// beside it: two traits embedded at the same depth would make CanHandle an
// ambiguous selector and the handler would stop satisfying [base.MessageHandler].
type PrivateCommandTrait struct {
	base.CommandHandlerTrait
}

func (*PrivateCommandTrait) GetScopes() []base.CommandScope {
	return []base.CommandScope{base.CommandScopeAllPrivateChats}
}

func (t *PrivateCommandTrait) CanHandle(reqenv *base.RequestEnv, msg *tgbotapi.Message) bool {
	return msg.Chat.IsPrivate() && t.CommandHandlerTrait.CanHandle(reqenv, msg)
}
