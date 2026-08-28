package dbmigrate

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/*.sql
var sqlFS embed.FS

// Apply runs bundled SQL (021+) on startup. Each file is independent: a failure
// is logged and the next file still runs, so PortalCustomerLinks can exist even
// if a later unique index in 021 fails on existing data.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.Glob(sqlFS, "sql/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(entries)
	var first error
	for _, name := range entries {
		body, err := sqlFS.ReadFile(name)
		if err != nil {
			return err
		}
		if err := execMulti(ctx, pool, string(body)); err != nil {
			fmt.Fprintf(os.Stderr, "dbmigrate %s: %v\n", name, err)
			if first == nil {
				first = fmt.Errorf("%s: %w", name, err)
			}
		}
	}
	return first
}

func execMulti(ctx context.Context, pool *pgxpool.Pool, sql string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Conn().PgConn().Exec(ctx, sql).ReadAll()
	return err
}
