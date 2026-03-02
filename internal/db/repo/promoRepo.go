package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/jackc/pgx/v5"

	"github.com/kozalosev/goSadTgBot/base"
)

type Promo struct {
	appEnv *base.ApplicationEnv
}

func NewPromo(appEnv *base.ApplicationEnv) *Promo {
	return &Promo{appEnv: appEnv}
}

func (p *Promo) CreatePromo(ctx context.Context, promoCode model.PromoCode) error {
	var db DBQuerier

	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		db = tx
	} else {
		db = p.appEnv.Database
	}

	const op = "Promo.CreatePromo"
	log := slog.With("op", op)

	query := `
		INSERT INTO Promo_codes
		(code, bonus_length, since, until, capacity)
		VALUES ($1, $2, $3, $4, $5)
		`

	_, err := db.Exec(
		ctx,
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
				slog.String("message", err.Error()),
				slog.String("component", "Database.Exec"),
				slog.String("promo_code", promoCode.Code)))
		return err
	}

	return nil
}

func (p *Promo) GetTable(ctx context.Context, codes []string) ([]model.ResponseCode, error) {
	const op = "Promo.GetTable"
	log := slog.With("op", op)
	var args []interface{}

	query := `
		SELECT code, bonus_length, capacity
		FROM promo_codes
		`

	if len(codes) > 0 {
		arg := make([]string, len(codes))
		for i, code := range codes {
			arg[i] = "%" + code + "%"
		}
		query += ` WHERE code ILIKE ANY($1)`
		args = append(args, arg)
	}

	query += `ORDER BY capacity;`
	rows, err := p.appEnv.Database.Query(ctx, query, args...)
	if err != nil {
		log.Error("failed to query promo codes table",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Database.Query")))
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
					slog.String("message", err.Error()),
					slog.String("component", "rows.Scan")))
			return nil, err
		}
		promo = append(promo, prom)
	}

	if err := rows.Err(); err != nil {
		log.Error("error iterating promo codes rows",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "rows.Err")))
		return nil, err
	}

	return promo, nil
}

func (p *Promo) GetPromoCode(ctx context.Context, codes []string) ([]model.StatResponseCode, error) {
	const op = "Promo.GetPromoCode"
	log := slog.With("op", op)

	if len(codes) == 0 {
		return nil, fmt.Errorf("%s, codes slice is empty", op)
	}

	query := `
	SELECT code, bonus_length, capacity,
		   count(uid) AS activations,
		   capacity + count(uid) AS initial_capacity
		FROM promo_codes
		JOIN promo_code_activations USING (code)
		WHERE code = any($1)
		GROUP BY code, bonus_length, capacity;
	`
	rows, err := p.appEnv.Database.Query(ctx, query, codes)
	if err != nil {
		log.Error("failed to query promo codes table",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Database.Query")))
		return nil, err
	}
	defer rows.Close()

	var promo []model.StatResponseCode

	for rows.Next() {
		var prom model.StatResponseCode
		err := rows.Scan(
			&prom.Code,
			&prom.BonusLength,
			&prom.Capacity,
			&prom.Activations,
			&prom.InitialCapacity,
		)
		if err != nil {
			log.Error("failed to scan promo code row",
				slog.Group("error",
					slog.String("message", err.Error()),
					slog.String("component", "rows.Scan")))
			return nil, err
		}
		promo = append(promo, prom)
	}

	if err := rows.Err(); err != nil {
		log.Error("error iterating promo codes rows",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "rows.Err")))
		return nil, err
	}

	return promo, nil

}
