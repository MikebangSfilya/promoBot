package repo

// реализация не под promoBot, пока можно сказать "нереальная" БД

import (
	"errors"
	"math/rand"

	"github.com/MikebangSfilya/promoBot/internal/db/dto"
	models "github.com/MikebangSfilya/promoBot/internal/model"
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

// Нужен для реализации интерфейса, в релаьном приложение не работает
func (service *UserService) FetchUserOptions(uid int64, defaultLang string) (settings.LangCode, settings.UserOptions) {
	var (
		language *string
		opts     dto.UserOptions
	)
	if err := service.appEnv.Database.QueryRow(service.appEnv.Ctx,
		"SELECT language, banned, role FROM Users WHERE uid = $1", uid).
		Scan(&language, &opts.Banned, &opts.Role); err != nil {
		// panic(err)
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

// Еще одна заглушка которая в будущем будет реально работать
func (service *UserService) CreatePromo(promoCode models.Promo) (bool, error) {

	query := `
	INSERT INTO Promo_codes
	(code, bonus_length, since, until, capacity)
	VALUES ($1, $2, $3, $4, $5)
	`

	if err := service.appEnv.Database.QueryRow(service.appEnv.Ctx, query); err != nil {

	}

	return true, nil
}
