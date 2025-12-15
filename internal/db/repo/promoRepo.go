package repo

import (
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/model"
	models "github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/kozalosev/goSadTgBot/base"
)

type Promo struct {
	appEnv *base.ApplicationEnv
}

func NewPromo(appEnv *base.ApplicationEnv) *Promo {
	return &Promo{appEnv: appEnv}
}

func (p *Promo) CreatePromo(promoCode models.PromoCode) error {
	const op = "promoRepo.sql.CreatePromo"
	query := `
	INSERT INTO Promo_codes
	(code, bonus_length, since, until, capacity)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := p.appEnv.Database.Exec(
		p.appEnv.Ctx,
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

func (p *Promo) GetTable() ([]model.ResponceCode, error) {
	query := `
	SELECT code, bonus_length, capacity
    FROM promo_codes
    ORDER BY capacity;
	`
	rows, err := p.appEnv.Database.Query(p.appEnv.Ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promo []model.ResponceCode

	for rows.Next() {
		var prom model.ResponceCode
		err := rows.Scan(
			&prom.Code,
			&prom.BonusLength,
			&prom.Capacity,
		)
		if err != nil {
			return nil, err
		}
		promo = append(promo, prom)
	}

	return promo, rows.Err()
}
