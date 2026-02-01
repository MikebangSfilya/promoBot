package repo

import (
	"context"
	"testing"
	"time"

	"github.com/MikebangSfilya/promoBot/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:14-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Apply migrations
	// Note: indentation inside the string below is manually aligned to tabs
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS Promo_Codes (
			code varchar(16) PRIMARY KEY,
			bonus_length integer NOT NULL,
			since date NOT NULL DEFAULT current_date,
			until date,
			capacity integer NOT NULL CHECK ( capacity >= 0 )
		);
	`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		require.NoError(t, postgresContainer.Terminate(ctx))
	}

	return pool, cleanup
}

func TestPromo_CreatePromo(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}

	repo := NewPromo(appEnv)

	tests := []struct {
		name    string
		promo   model.PromoCode
		wantErr bool
	}{
		{
			name: "successful creation",
			promo: model.PromoCode{
				Code:        "TEST123",
				BonusLength: 10,
				Since:       time.Now(),
				Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
				Capacity:    5,
			},
			wantErr: false,
		},
		{
			name: "promo with nil until",
			promo: model.PromoCode{
				Code:        "TEST456",
				BonusLength: 20,
				Since:       time.Now(),
				Until:       nil,
				Capacity:    10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreatePromo(ctx, tt.promo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify that the promo code was actually created
				var count int
				err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Codes WHERE code = $1", tt.promo.Code).Scan(&count)
				require.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestPromo_GetTable(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}

	repo := NewPromo(appEnv)

	// Create test data
	testPromos := []model.PromoCode{
		{
			Code:        "PROMO1",
			BonusLength: 5,
			Since:       time.Now(),
			Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
			Capacity:    10,
		},
		{
			Code:        "PROMO2",
			BonusLength: 15,
			Since:       time.Now(),
			Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
			Capacity:    5,
		},
		{
			Code:        "PROMO3",
			BonusLength: 20,
			Since:       time.Now(),
			Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
			Capacity:    15,
		},
	}

	// Create promo codes via repository
	for _, promo := range testPromos {
		err := repo.CreatePromo(ctx, promo)
		require.NoError(t, err)
	}

	// Retrieve the table
	result, err := repo.GetTable(ctx)
	require.NoError(t, err)

	// Verify that we received all promo codes
	assert.Len(t, result, 3)

	// Check sorting by capacity (should be ascending)
	assert.Equal(t, "PROMO2", result[0].Code) // capacity = 5
	assert.Equal(t, "PROMO1", result[1].Code) // capacity = 10
	assert.Equal(t, "PROMO3", result[2].Code) // capacity = 15

	// Verify data
	assert.Equal(t, "PROMO2", result[0].Code)
	assert.Equal(t, 15, result[0].BonusLength)
	assert.Equal(t, 5, result[0].Capacity)

	assert.Equal(t, "PROMO1", result[1].Code)
	assert.Equal(t, 5, result[1].BonusLength)
	assert.Equal(t, 10, result[1].Capacity)

	assert.Equal(t, "PROMO3", result[2].Code)
	assert.Equal(t, 20, result[2].BonusLength)
	assert.Equal(t, 15, result[2].Capacity)
}

func TestPromo_GetTable_Empty(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}

	repo := NewPromo(appEnv)

	// Get table from empty DB
	result, err := repo.GetTable(ctx)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestPromo_CreatePromo_Duplicate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}

	repo := NewPromo(appEnv)

	promo := model.PromoCode{
		Code:        "DUPLICATE",
		BonusLength: 10,
		Since:       time.Now(),
		Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
		Capacity:    5,
	}

	// Create the first promo code
	err := repo.CreatePromo(ctx, promo)
	require.NoError(t, err)

	// Try to create a duplicate
	err = repo.CreatePromo(ctx, promo)
	assert.Error(t, err) // Expect an error due to PRIMARY KEY violation
}

func TestPromo_CreatePromo_InTransaction(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}
	repo := NewPromo(appEnv)

	promo := model.PromoCode{
		Code:        "TX_PROMO",
		BonusLength: 10,
		Since:       time.Now(),
		Capacity:    5,
	}

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	defer func(tx pgx.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	ctxWithTx := context.WithValue(ctx, TxKey{}, tx)

	err = repo.CreatePromo(ctxWithTx, promo)
	require.NoError(t, err)

	// Verify INSIDE transaction (should be visible)
	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Codes WHERE code = $1", promo.Code).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Row should be visible inside transaction")

	// Verify OUTSIDE transaction (should NOT be visible before commit)
	var countOutside int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Codes WHERE code = $1", promo.Code).Scan(&countOutside)
	require.NoError(t, err)
	assert.Equal(t, 0, countOutside, "Row should not be visible outside transaction before commit")

	// Commit transaction
	err = tx.Commit(ctx)
	require.NoError(t, err)

	// Verify OUTSIDE transaction after commit (should be visible)
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Codes WHERE code = $1", promo.Code).Scan(&countOutside)
	require.NoError(t, err)
	assert.Equal(t, 1, countOutside, "Row should be visible after commit")
}
