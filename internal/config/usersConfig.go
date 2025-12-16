package config

import (
	"os"

	"github.com/kozalosev/goSadTgBot/settings"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type UsersConfig struct {
	users []User
}

func NewUsersConfig() *UsersConfig {
	file, err := os.ReadFile("/app/users.yaml")
	if err != nil {
		panic("failed to read users.yaml file")
	}

	var cfg struct {
		Users []User
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		panic("failed to parse users.yaml file")
	}
	return &UsersConfig{users: cfg.Users}
}

func (u *UsersConfig) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, settings.UserOptions) {
	userId := UserId(uid)
	user, found := lo.Find(u.users, func(user User) bool { return user.UID == userId })
	if !found {
		return settings.LangCode(defaultLang), UserOptions{}
	}
	return settings.LangCode(user.Language), UserOptions{Role: user.Role}
}
