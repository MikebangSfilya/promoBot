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
	const op = "Promo.CreatePromo"
	log := slog.With("op", op)

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
		log.Error("failed to create promo code",
			slog.Group("error",
				"message", err.Error(),
				"component", "Database.Exec",
				"promo_code", promoCode.Code))
		return err
	}

	return nil
}

func (p *Promo) GetTable() ([]model.ResponseCode, error) {
	const op = "Promo.GetTable"
	log := slog.With("op", op)

	query := `
		SELECT code, bonus_length, capacity
		FROM promo_codes
		ORDER BY capacity;
		`
	rows, err := p.appEnv.Database.Query(p.appEnv.Ctx, query)
	if err != nil {
		log.Error("failed to query promo codes table",
			slog.Group("error",
				"message", err.Error(),
				"component", "Database.Query"))
		return nil, err
	}
	defer rows.Close()

	var promo []model.ResponseCode

	for rows.Next() {
		var prom model.ResponseCode
		err := rows.Scan(
			&prom.Code,
			&prom.BonusLength,
			&prom.Capacity,
		)
		if err != nil {
			log.Error("failed to scan promo code row",
				slog.Group("error",
					"message", err.Error(),
					"component", "rows.Scan"))
			return nil, err
		}
		promo = append(promo, prom)
	}

	if err := rows.Err(); err != nil {
		log.Error("error iterating promo codes rows",
			slog.Group("error",
				"message", err.Error(),
				"component", "rows.Err"))
		return nil, err
	}

	return promo, nil
}
