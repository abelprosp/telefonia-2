package dbmigrate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var sqlFS embed.FS

// Apply runs bundled SQL in order, committing each file before the next one.
// A failed migration prevents startup with an incompatible schema.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.Glob(sqlFS, "sql/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(entries)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Only one API process may bootstrap the schema at a time. This matters
	// during rolling deploys, when the old and new containers overlap.
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(hashtext('luxus-connect-schema'))`); err != nil {
		return err
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock(hashtext('luxus-connect-schema'))`)
	}()

	for _, name := range entries {
		body, err := sqlFS.ReadFile(name)
		if err != nil {
			return err
		}
		if err := execMulti(ctx, conn, string(body)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func execMulti(ctx context.Context, conn *pgxpool.Conn, sql string) error {
	_, err := conn.Conn().PgConn().Exec(ctx, sql).ReadAll()
	return err
}
