package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type LocalUser struct {
	Version          int64    `json:"-"`
	ID               string   `json:"id"`
	Username         string   `json:"username"`
	Email            string   `json:"email"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	OrganizationID   string   `json:"organization_id"`
	OrganizationName string   `json:"organization_name"`
	Roles            []string `json:"roles"`
	Document         string   `json:"document"`
	Enabled          bool     `json:"enabled"`
	PasswordHash     string   `json:"-"`
	PasskeyCount     int      `json:"-"`
}

const localColumns = `id,username,email,first_name,last_name,organization_id,organization_name,roles,document,enabled,password_hash,version,(SELECT count(*)::int FROM auth_webauthn_credentials c WHERE c.user_id=auth_users.id) AS passkey_count`

func scanLocal(row pgx.Row) (*LocalUser, error) {
	u := &LocalUser{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.FirstName, &u.LastName, &u.OrganizationID, &u.OrganizationName, &u.Roles, &u.Document, &u.Enabled, &u.PasswordHash, &u.Version, &u.PasskeyCount)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}
func (s *Store) LocalUserByID(ctx context.Context, id string) (*LocalUser, error) {
	return scanLocal(s.pool.QueryRow(ctx, `SELECT `+localColumns+` FROM auth_users WHERE id=$1`, id))
}
func (s *Store) LocalUserByUsername(ctx context.Context, name string) (*LocalUser, error) {
	return scanLocal(s.pool.QueryRow(ctx, `SELECT `+localColumns+` FROM auth_users WHERE lower(username)=lower($1)`, name))
}
func (s *Store) ListLocalUsers(ctx context.Context, org, search string) ([]LocalUser, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+localColumns+` FROM auth_users WHERE organization_id=$1 AND (username ILIKE '%'||$2||'%' OR email ILIKE '%'||$2||'%') ORDER BY username LIMIT 200`, org, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []LocalUser{}
	for rows.Next() {
		u, e := scanLocal(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}
func (s *Store) CreateLocalUser(ctx context.Context, u *LocalUser) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO auth_users (id,username,email,first_name,last_name,organization_id,organization_name,roles,document,enabled,password_hash) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, u.ID, u.Username, u.Email, u.FirstName, u.LastName, u.OrganizationID, u.OrganizationName, u.Roles, u.Document, u.Enabled, u.PasswordHash)
	return err
}

// Every account change invalidates its sessions in the same transaction.
func (s *Store) UpdateLocalUser(ctx context.Context, u *LocalUser) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE auth_users SET email=$2,first_name=$3,last_name=$4,roles=$5,enabled=$6,password_hash=$7,organization_name=$8,organization_id=$9,version=version+1 WHERE id=$1 AND version=$10`, u.ID, u.Email, u.FirstName, u.LastName, u.Roles, u.Enabled, u.PasswordHash, u.OrganizationName, u.OrganizationID, u.Version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("account changed concurrently; reload and retry")
	}
	if _, err = tx.Exec(ctx, `DELETE FROM auth_sessions WHERE user_id=$1`, u.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
