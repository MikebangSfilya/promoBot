package handlers

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/kozalosev/goSadTgBot/base"
)

const groupCommandsDisabled = "groupCommandsDisabled"

// GroupGuardHandler swallows everything the bot receives in a group chat.
//
// Without it the framework's fallback would answer every unmatched group
// message with "Unknown command", and a group message sent by someone with an
// open wizard in DM would be consumed as an answer to that form. It is a plain
// [base.MessageHandler] rather than a CommandHandler so it never shows up in
// the command menu, and it must be registered last, since the first handler
// whose CanHandle returns true wins.
type GroupGuardHandler struct {
	appEnv *base.ApplicationEnv
}

func NewGroupGuardHandler(appEnv *base.ApplicationEnv) *GroupGuardHandler {
	return &GroupGuardHandler{appEnv: appEnv}
}

func (*GroupGuardHandler) CanHandle(_ *base.RequestEnv, msg *tgbotapi.Message) bool {
	return !msg.Chat.IsPrivate()
}

// Handle tells the user where the commands live, but only when they actually
// tried to run one — ordinary group chatter is left alone.
func (h *GroupGuardHandler) Handle(reqEnv *base.RequestEnv, msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return
	}
	base.NewReplier(h.appEnv, reqEnv, msg)(groupCommandsDisabled)
}
