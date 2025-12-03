package repo

// реализация не под promoBot, пока можно сказать "нереальная" БД

import (
	"errors"
	"log/slog"
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
 
func (service *UserService) ChangeLanguage(uid int64, lang settings.LangCode) error {
	a := 8
	if rand.Intn(10) > a {
		err := errors.New("fail")
		return err
	}

	return nil
}


func (service *UserService) CreatePromo(promoCode models.PromoCode) error {
	const op = "promoRepo.sql.CreatePromo"
	query := `
	INSERT INTO Promo_codes
	(code, bonus_length, since, until, capacity)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := service.appEnv.Database.Exec(
		service.appEnv.Ctx,
		query,
		promoCode.Code,
		promoCode.BonusLength,
		promoCode.Since,
		promoCode.Until,
		promoCode.Capacity,
	)
	if err != nil {
		slog.Error("faield to Exec", "errror", err)
		return err
	}

	return nil
}
