package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

// Administrative recovery requires shell/container access, never a public API key.
// Passwords arrive via stdin so they do not appear in argv or Docker inspect.
func runAuthCommand(ctx context.Context, st *store.Store, cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: api auth init-admin USUARIO EMAIL | reset-password USUARIO | reset-mfa|reset-passkeys USUARIO | import-users ARQUIVO.json")
	}
	if args[0] == "import-users" {
		if len(args) < 2 {
			return fmt.Errorf("informe o arquivo JSON ou - para stdin")
		}
		assignments := []string{}
		fallback := ""
		for i := 2; i < len(args); i++ {
			if (args[i] != "--assign-org" && args[i] != "--assign-unassigned-to") || i+1 >= len(args) {
				return fmt.Errorf("uso: api auth import-users ARQUIVO.json [--assign-org USUARIO=UUID] [--assign-unassigned-to UUID]")
			}
			option := args[i]
			i++
			if option == "--assign-unassigned-to" {
				if fallback != "" {
					return fmt.Errorf("informe --assign-unassigned-to apenas uma vez")
				}
				fallback = args[i]
			} else {
				assignments = append(assignments, args[i])
			}
		}
		return importAccountsFile(ctx, st, args[1], fallback, assignments...)
	}
	if len(args) < 2 {
		return fmt.Errorf("informe o usuário")
	}
	username := strings.TrimSpace(args[1])
	u, err := st.LocalUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	switch args[0] {
	case "init-admin":
		if len(args) != 3 {
			return fmt.Errorf("uso: api auth init-admin USUARIO EMAIL")
		}
		var exists bool
		if err := st.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth_users WHERE enabled AND roles && ARRAY['master','admin'])`).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("já existe um administrador; use reset-password para recuperar o acesso")
		}
		if u != nil {
			return fmt.Errorf("usuário já existe; use reset-password")
		}
		u = &store.LocalUser{ID: uuid.NewString(), Username: username, Email: strings.TrimSpace(args[2]), FirstName: "Administrador", OrganizationID: auth.DefaultLuxusOrganizationID, OrganizationName: "Luxus Connect", Enabled: true, Roles: []string{"master", "user"}}
	case "reset-password", "reset-mfa", "reset-passkeys":
		if u == nil {
			return fmt.Errorf("usuário não encontrado")
		}
	default:
		return fmt.Errorf("comando desconhecido")
	}
	if args[0] == "reset-mfa" || args[0] == "reset-passkeys" {
		tx, e := st.Pool().Begin(ctx)
		if e != nil {
			return e
		}
		defer tx.Rollback(ctx)
		// Clear legacy TOTP state without dropping the columns, and remove every passkey.
		if _, e = tx.Exec(ctx, `UPDATE auth_users SET mfa_secret='',mfa_last_step=-1,version=version+1 WHERE id=$1`, u.ID); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `DELETE FROM auth_webauthn_credentials WHERE user_id=$1`, u.ID); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `DELETE FROM auth_webauthn_challenges WHERE user_id=$1`, u.ID); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `DELETE FROM auth_recovery_codes WHERE user_id=$1`, u.ID); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `DELETE FROM auth_sessions WHERE user_id=$1`, u.ID); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `DELETE FROM auth_attempts WHERE bucket=ANY($1)`, localauth.AccountAttemptBuckets(u)); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `INSERT INTO auth_events(user_id,event)VALUES($1,'admin_mfa_reset')`, u.ID); e != nil {
			return e
		}
		if e = tx.Commit(ctx); e != nil {
			return e
		}
		fmt.Println("Passkeys redefinidas. Todas as sessões foram encerradas.")
		return nil
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 258))
	if err != nil {
		return err
	}
	password := strings.TrimSuffix(strings.TrimSuffix(string(raw), "\n"), "\r")
	hash, err := localauth.HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	if args[0] == "init-admin" {
		err = st.CreateLocalUser(ctx, u)
	} else {
		err = st.UpdateLocalUser(ctx, u)
	}
	if err != nil {
		return err
	}
	if _, err = st.Pool().Exec(ctx, `DELETE FROM auth_attempts WHERE bucket=ANY($1)`, localauth.AccountAttemptBuckets(u)); err != nil {
		return err
	}
	st.Pool().Exec(ctx, `INSERT INTO auth_events(user_id,event)VALUES($1,'admin_password_set')`, u.ID)
	fmt.Println("Senha configurada. Nenhuma credencial foi exibida.")
	return nil
}

func parseOrganizationAssignments(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, value := range values {
		name, id, ok := strings.Cut(value, "=")
		name = strings.TrimSpace(name)
		id = strings.TrimSpace(id)
		parsed, err := uuid.Parse(id)
		if !ok || name == "" || err != nil || parsed == uuid.Nil {
			return nil, fmt.Errorf("vínculo inválido; use --assign-org USUARIO=UUID")
		}
		if _, exists := out[name]; exists {
			return nil, fmt.Errorf("usuário %s informado mais de uma vez", name)
		}
		out[name] = parsed.String()
	}
	return out, nil
}
