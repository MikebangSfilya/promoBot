package common

import "github.com/kozalosev/goSadTgBot/base"

type GroupCommand interface {
	base.CommandHandler

	isGroupCommand() bool
}

var CommandScopePrivateAndGroupChats = []base.CommandScope{base.CommandScopeAllPrivateChats, base.CommandScopeAllGroupChats, base.CommandScopeAllChatAdmins}

type GroupCommandTrait struct{}

//nolint:unused
func (*GroupCommandTrait) isGroupCommand() bool {
	return true
}

func (*GroupCommandTrait) GetScopes() []base.CommandScope {
	return CommandScopePrivateAndGroupChats
}
