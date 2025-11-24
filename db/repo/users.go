package repo

// реализация не под promoBot, пока можно сказать "нереальная" БД

import (
	"errors"
	"math/rand"

	"github.com/MikebangSfilya/promoBot/db/dto"
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

// FetchUserOptions is the implementation of the [settings.OptionsFetcher.FetchUserOptions] method for this application.
func (service *UserService) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, settings.UserOptions) {
	var (
		language *string
		opts     dto.UserOptions
	)
	if err := service.appEnv.Database.QueryRow(service.appEnv.Ctx,
		"SELECT language, banned, role FROM Users WHERE uid = $1", uid).
		Scan(&language, &opts.Banned, &opts.Role); err != nil {

	}
	if language == nil {
		language = &defaultLang
	}
	return settings.LangCode(*language), &opts
}

// Заглушка для работы со сменой языка
func (service *UserService) ChangeLanguage(uid int64, lang settings.LangCode) error {
	a := 8
	if rand.Intn(10) > a {
		err := errors.New("fail")
		return err
	}

	return nil
}
