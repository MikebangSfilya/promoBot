package repo

import (
	"context"
	"os"
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
	testcontainers.SkipIfProviderIsNotHealthy(t)

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

	// Apply migrations from the actual migration file
	migration, err := os.ReadFile("../../../db/migrations/000001_create_tables.up.sql")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(migration))
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
		name         string
		promo        model.PromoCode
		wantErr      bool
		validateFunc func(t *testing.T, row model.PromoCode)
	}{
		{
			name: "successful creation",
			promo: model.PromoCode{
				Code:        "TEST123",
				BonusLength: 10,
				Since:       func() *time.Time { t := time.Now(); return &t }(),
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
				Since:       func() *time.Time { t := time.Now(); return &t }(),
				Until:       nil,
				Capacity:    10,
			},
			wantErr: false,
		},
		{
			name: "promo with nil since uses current_date, not 0001-01-01",
			promo: model.PromoCode{
				Code:        "TEST789",
				BonusLength: 5,
				Since:       nil,
				Until:       nil,
				Capacity:    3,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, row model.PromoCode) {
				require.NotNil(t, row.Since)
				assert.False(t, row.Since.IsZero(), "since must not be the zero time (0001-01-01)")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreatePromo(ctx, tt.promo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				rows, err := pool.Query(ctx,
					"SELECT code, bonus_length, since, until, capacity FROM Promo_Codes WHERE code = $1",
					tt.promo.Code)
				require.NoError(t, err)
				defer rows.Close()

				var fetched []model.PromoCode
				for rows.Next() {
					var row model.PromoCode
					require.NoError(t, rows.Scan(&row.Code, &row.BonusLength, &row.Since, &row.Until, &row.Capacity))
					fetched = append(fetched, row)
				}
				require.NoError(t, rows.Err())
				require.Len(t, fetched, 1)

				if tt.validateFunc != nil {
					tt.validateFunc(t, fetched[0])
				}
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
			Since:       func() *time.Time { t := time.Now(); return &t }(),
			Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
			Capacity:    10,
		},
		{
			Code:        "PROMO2",
			BonusLength: 15,
			Since:       func() *time.Time { t := time.Now(); return &t }(),
			Until:       func() *time.Time { t := time.Now().Add(30 * 24 * time.Hour); return &t }(),
			Capacity:    5,
		},
		{
			Code:        "PROMO3",
			BonusLength: 20,
			Since:       func() *time.Time { t := time.Now(); return &t }(),
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

	// Check filtration also works
	result, err = repo.GetTable(ctx, "1")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "PROMO1", result[0].Code)
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
		Since:       func() *time.Time { t := time.Now(); return &t }(),
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
		Since:       func() *time.Time { t := time.Now(); return &t }(),
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

func TestPromo_UpdatePromo(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}
	repo := NewPromo(appEnv)

	promo := model.PromoCode{
		Code:        "UPD_PROMO",
		BonusLength: 10,
		Capacity:    5,
	}
	require.NoError(t, repo.CreatePromo(ctx, promo))

	until := time.Now().Add(30 * 24 * time.Hour)
	updated := model.PromoCode{
		Code:        promo.Code,
		BonusLength: 20,
		Capacity:    15,
		Until:       &until,
	}
	require.NoError(t, repo.UpdatePromo(ctx, updated))

	var fetched model.PromoCode
	err := pool.QueryRow(ctx,
		"SELECT code, bonus_length, since, until, capacity FROM Promo_Codes WHERE code = $1",
		promo.Code,
	).Scan(&fetched.Code, &fetched.BonusLength, &fetched.Since, &fetched.Until, &fetched.Capacity)
	require.NoError(t, err)

	assert.Equal(t, updated.Code, fetched.Code)
	assert.Equal(t, updated.BonusLength, fetched.BonusLength)
	assert.Equal(t, updated.Capacity, fetched.Capacity)
	require.NotNil(t, fetched.Until)
	assert.False(t, fetched.Until.IsZero())
}

func TestPromo_DeletePromo(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	appEnv := &base.ApplicationEnv{
		Database: pool,
		Ctx:      ctx,
	}
	repo := NewPromo(appEnv)

	promo := model.PromoCode{
		Code:        "DEL_PROMO",
		BonusLength: 10,
		Capacity:    5,
	}
	require.NoError(t, repo.CreatePromo(ctx, promo))
	_, err := pool.Exec(ctx,
		"INSERT INTO Promo_Code_Activations (uid, code, affected_chats) VALUES ($1, $2, $3)",
		int64(1),
		promo.Code,
		1,
	)
	require.NoError(t, err)

	require.NoError(t, repo.DeletePromo(ctx, promo.Code))

	var promoCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Codes WHERE code = $1", promo.Code).Scan(&promoCount)
	require.NoError(t, err)
	assert.Equal(t, 0, promoCount)

	var activationCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM Promo_Code_Activations WHERE code = $1", promo.Code).Scan(&activationCount)
	require.NoError(t, err)
	assert.Equal(t, 0, activationCount)
}
