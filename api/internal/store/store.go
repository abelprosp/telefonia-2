package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxus-connect/telefonia/api/internal/dbmigrate"
)

type Store struct {
	pool *pgxpool.Pool
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	s := &Store{pool: pool}
	if err := dbmigrate.Apply(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("dbmigrate: %w", err)
	}
	if err := s.ensureOrganizationSettingsSchema(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ensure OrganizationSettings schema: %v\n", err)
	}
	if err := s.ensureProcessingMonthSchema(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ensure ProcessingMonths schema: %v\n", err)
	}
	if err := s.ensureFinancialAgentSchema(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ensure FinancialAgent schema: %v\n", err)
	}
	return s, nil
}

func isUndefinedColumn(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42703"
}

func (s *Store) ensureOrganizationSettingsSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		ALTER TABLE IF EXISTS "OrganizationSettings"
		    ADD COLUMN IF NOT EXISTS "ProrataDivisor" integer NOT NULL DEFAULT 30,
		    ADD COLUMN IF NOT EXISTS "SicrediEnabled" boolean NOT NULL DEFAULT FALSE,
		    ADD COLUMN IF NOT EXISTS "SicrediSandbox" boolean NOT NULL DEFAULT TRUE,
		    ADD COLUMN IF NOT EXISTS "SicrediAPIKey" text NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediUsername" character varying(128) NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediPassword" text NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediCooperativa" character varying(32) NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediPosto" character varying(32) NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediCodigoBeneficiario" character varying(32) NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediAccountNumber" character varying(64) NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediWebhookToken" text NOT NULL DEFAULT '',
		    ADD COLUMN IF NOT EXISTS "SicrediPublicAPIURL" text NOT NULL DEFAULT ''`)
	return err
}

func (s *Store) ensureFinancialAgentSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS "FinancialAgentSettings" (
			"OrganizationId" character varying(36) NOT NULL,
			"Enabled" boolean NOT NULL DEFAULT true,
			"EvolutionApiUrl" text NOT NULL DEFAULT '',
			"EvolutionApiKey" text NOT NULL DEFAULT '',
			"EvolutionInstance" character varying(128) NOT NULL DEFAULT 'luxus',
			"N8nWebhookUrl" text NOT NULL DEFAULT '',
			"UpdatedAt" timestamp with time zone NOT NULL DEFAULT now(),
			"UpdatedBy" character varying(256),
			CONSTRAINT "PK_FinancialAgentSettings" PRIMARY KEY ("OrganizationId")
		);
		CREATE TABLE IF NOT EXISTS "FinancialAgentEvents" (
			"Id" character varying(36) NOT NULL,
			"OrganizationId" character varying(36) NOT NULL,
			"EventType" character varying(64) NOT NULL,
			"WhatsAppNumber" character varying(32),
			"CustomerId" character varying(36),
			"CustomerName" character varying(256),
			"InvoiceId" character varying(36),
			"InvoiceNumber" character varying(64),
			"Success" boolean NOT NULL DEFAULT true,
			"Summary" text NOT NULL DEFAULT '',
			"CreatedAt" timestamp with time zone NOT NULL DEFAULT now(),
			CONSTRAINT "PK_FinancialAgentEvents" PRIMARY KEY ("Id")
		);
		CREATE INDEX IF NOT EXISTS "IX_FinancialAgentEvents_Org_Created"
			ON "FinancialAgentEvents" ("OrganizationId", "CreatedAt" DESC)`)
	return err
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) WithTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type txCtxKey struct{}

func CtxWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func (s *Store) q(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok {
		return tx
	}
	return s.pool
}
