package repo

import (
	"errors"

	"github.com/MikebangSfilya/promoBot/internal/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/settings"
)

var NoRowsWereAffected = errors.New("no rows were affected")

// UserService is a repository for the Users table.
type UserService struct {
	appEnv *base.ApplicationEnv
}

func NewUserService(appEnv *base.ApplicationEnv) *UserService {
	return &UserService{appEnv: appEnv}
}

func (service *UserService) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, settings.UserOptions) {
	var (
		language *string
		opts     dto.UserOptions
	)
	if err := service.appEnv.Database.QueryRow(service.appEnv.Ctx,
		"SELECT language, banned, role FROM Users_promo WHERE uid = $1", uid).
		Scan(&language, &opts.Banned, &opts.Role); err != nil {
	}
	if language == nil {
		language = &defaultLang
	}
	return settings.LangCode(*language), &opts
}

func (service *UserService) ChangeLanguage(uid int64, lang settings.LangCode) error {
	_, err := service.appEnv.Database.Exec(service.appEnv.Ctx,
		"UPDATE Users_promo SET language = $2 WHERE uid = $1", uid, lang)
	return err
}
