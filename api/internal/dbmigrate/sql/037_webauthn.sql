CREATE TABLE IF NOT EXISTS auth_webauthn_credentials (
  credential_id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
  rp_id text NOT NULL,
  name text NOT NULL DEFAULT '',
  credential jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  last_used_at timestamptz
);
CREATE INDEX IF NOT EXISTS auth_webauthn_credentials_user ON auth_webauthn_credentials(user_id);
CREATE TABLE IF NOT EXISTS auth_webauthn_challenges (
  session_hash text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('register','login')),
  user_id text NOT NULL REFERENCES auth_users(id) ON DELETE CASCADE,
  session_data jsonb NOT NULL,
  expires_at timestamptz NOT NULL,
  PRIMARY KEY (session_hash, kind)
);
CREATE INDEX IF NOT EXISTS auth_webauthn_challenges_expires ON auth_webauthn_challenges(expires_at);
