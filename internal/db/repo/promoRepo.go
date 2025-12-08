package repo

import (
	"log/slog"

	models "github.com/MikebangSfilya/promoBot/internal/model"
)

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
