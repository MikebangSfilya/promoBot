package common

import "github.com/kozalosev/goSadTgBot/base"

type PrivateCommandTrait struct{}

func (*PrivateCommandTrait) GetScopes() []base.CommandScope {
	return []base.CommandScope{base.CommandScopeAllPrivateChats}
}
