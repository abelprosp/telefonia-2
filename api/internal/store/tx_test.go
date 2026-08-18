package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		raw = "postgres://postgres:postgres@127.0.0.1:5432/luxus_connect_dev?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	st, err := store.New(ctx, config.NormalizeDatabaseURL(raw))
	if err != nil {
		t.Skipf("postgres indisponível para teste de transação: %v", err)
	}
	t.Cleanup(st.Close)
	return st
}

func TestWithTx_FailedImportRollsBack(t *testing.T) {
	st := testStore(t)
	ctx := context.Background()
	id := uuid.New().String()

	err := st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = store.CtxWithTx(ctx, tx)
		if _, err := st.Pool().Exec(ctx, `SELECT 1`); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `CREATE TEMP TABLE IF NOT EXISTS tx_rollback_probe (id text PRIMARY KEY)`)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tx_rollback_probe (id) VALUES ($1)`, id); err != nil {
			return err
		}
		return pgx.ErrTxClosed
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	var exists bool
	_ = st.Pool().QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'tx_rollback_probe')`).Scan(&exists)
	if exists {
		var n int
		_ = st.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM tx_rollback_probe WHERE id = $1`, id).Scan(&n)
		if n != 0 {
			t.Fatalf("row should have rolled back, found %d", n)
		}
	}
}
