package repo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"

	"github.com/kozalosev/goSadTgBot/base"
)

type Promo struct {
	appEnv *base.ApplicationEnv
}

func NewPromo(appEnv *base.ApplicationEnv) *Promo {
	return &Promo{appEnv: appEnv}
}

func (p *Promo) dbFromContext(ctx context.Context) DBQuerier {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx
	}
	return p.appEnv.Database
}

func (p *Promo) CreatePromo(ctx context.Context, promoCode model.PromoCode) error {
	const op = "Promo.CreatePromo"
	log := slog.With("op", op)
	db := p.dbFromContext(ctx)

	query := `
		INSERT INTO Promo_codes
		(code, bonus_length, since, until, capacity)
		VALUES ($1, $2, COALESCE($3, current_date), $4, $5)
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

func (p *Promo) UpdatePromo(ctx context.Context, promoCode model.PromoCode) error {
	const op = "Promo.UpdatePromo"
	log := slog.With("op", op)
	db := p.dbFromContext(ctx)

	query := `
		UPDATE promo_codes
		SET bonus_length = $2,
			since = COALESCE($3, current_date),
			until = $4,
			capacity = $5
		WHERE code = $1
	`

	tag, err := db.Exec(
		ctx,
		query,
		promoCode.Code,
		promoCode.BonusLength,
		promoCode.Since,
		promoCode.Until,
		promoCode.Capacity,
	)
	if err != nil {
		log.Error("failed to update promo code",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Database.Exec"),
				slog.String("promo_code", promoCode.Code)))
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s, promo code not found: %s", op, promoCode.Code)
	}

	return nil
}

func (p *Promo) DeletePromo(ctx context.Context, code string) error {
	const op = "Promo.DeletePromo"
	log := slog.With("op", op)
	db := p.dbFromContext(ctx)

	if _, err := db.Exec(ctx, "DELETE FROM promo_code_activations WHERE code = $1", code); err != nil {
		log.Error("failed to delete promo code activations",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Database.Exec"),
				slog.String("promo_code", code)))
		return err
	}

	tag, err := db.Exec(ctx, "DELETE FROM promo_codes WHERE code = $1", code)
	if err != nil {
		log.Error("failed to delete promo code",
			slog.Group("error",
				slog.String("message", err.Error()),
				slog.String("component", "Database.Exec"),
				slog.String("promo_code", code)))
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s, promo code not found: %s", op, code)
	}

	return nil
}

func (p *Promo) GetTable(ctx context.Context, codes ...string) ([]model.ResponseCode, error) {
	const op = "Promo.GetTable"
	log := slog.With("op", op)

	query := `
		SELECT code, bonus_length, capacity
		FROM promo_codes
		WHERE cardinality($1::text[]) = 0 OR code ILIKE ANY($1)
		ORDER BY capacity;
	`

	args := lo.Map(codes, func(code string, _ int) string {
		return "%" + code + "%"
	})

	rows, err := p.appEnv.Database.Query(ctx, query, args)
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

func (p *Promo) GetPromoCode(ctx context.Context, codes ...string) ([]model.StatResponseCode, error) {
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
