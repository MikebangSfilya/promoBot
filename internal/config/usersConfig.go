package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/kozalosev/goSadTgBot/settings"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type UsersConfig struct {
	users []User
}

func NewUsersConfig(configPath string) (*UsersConfig, error) {
	if configPath == "" {
		return nil, errors.New("users configuration file path is not specified")
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read users.yaml file: %w", err)
	}

	var cfg struct {
		Users []User
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse users.yaml file: %w", err)
	}
	return &UsersConfig{users: cfg.Users}, nil
}

func (u *UsersConfig) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, settings.UserOptions) {
	userId := UserId(uid)
	user, found := lo.Find(u.users, func(user User) bool { return user.UID == userId })
	if !found {
		return settings.LangCode(defaultLang), UserOptions{}
	}
	return settings.LangCode(user.Language), UserOptions{Role: user.Role, UserName: user.Name}
}
