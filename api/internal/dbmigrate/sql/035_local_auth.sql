CREATE TABLE IF NOT EXISTS auth_users (
 id text PRIMARY KEY,
 username text NOT NULL,
 email text NOT NULL,
 first_name text NOT NULL DEFAULT '', last_name text NOT NULL DEFAULT '',
 organization_id text NOT NULL, organization_name text NOT NULL DEFAULT '',
 roles text[] NOT NULL DEFAULT '{user}', document text NOT NULL DEFAULT '',
 enabled boolean NOT NULL DEFAULT true,
 password_hash text NOT NULL DEFAULT '',
 mfa_secret text NOT NULL DEFAULT '', mfa_last_step bigint NOT NULL DEFAULT -1,
 version bigint NOT NULL DEFAULT 1,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS auth_users_username ON auth_users (lower(username));
CREATE TABLE IF NOT EXISTS auth_sessions (
 token_hash text PRIMARY KEY, user_id text NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
 mfa_verified boolean NOT NULL DEFAULT false,
 pending_secret text NOT NULL DEFAULT '', pending_expires timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(), last_seen timestamptz NOT NULL DEFAULT now(),
 expires_at timestamptz NOT NULL, user_version bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS auth_sessions_user ON auth_sessions(user_id);
CREATE TABLE IF NOT EXISTS auth_recovery_codes (
 user_id text NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
 code_hash text NOT NULL, PRIMARY KEY(user_id, code_hash)
);
CREATE TABLE IF NOT EXISTS auth_attempts (
 bucket text PRIMARY KEY, attempts integer NOT NULL, expires_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS auth_events (
 id bigserial PRIMARY KEY, user_id text, event text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
