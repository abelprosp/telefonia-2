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
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

// Administrative recovery requires shell/container access, never a public API key.
// Passwords arrive via stdin so they do not appear in argv or Docker inspect.
func runAuthCommand(ctx context.Context, st *store.Store, cfg config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("uso: api auth init-admin USUARIO EMAIL | reset-password USUARIO | reset-mfa USUARIO | import-keycloak")
	}
	if args[0] == "import-keycloak" {
		assignments := []string{}
		fallback := ""
		for i := 1; i < len(args); i++ {
			if (args[i] != "--assign-org" && args[i] != "--assign-unassigned-to") || i+1 >= len(args) {
				return fmt.Errorf("uso: api auth import-keycloak [--assign-org USUARIO=UUID] [--assign-unassigned-to UUID]")
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
		return importLegacyAccountsWithDefault(ctx, st, cfg, fallback, assignments...)
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
	case "reset-password", "reset-mfa":
		if u == nil {
			return fmt.Errorf("usuário não encontrado")
		}
	default:
		return fmt.Errorf("comando desconhecido")
	}
	if args[0] == "reset-mfa" {
		tx, e := st.Pool().Begin(ctx)
		if e != nil {
			return e
		}
		defer tx.Rollback(ctx)
		if _, e = tx.Exec(ctx, `UPDATE auth_users SET mfa_secret='',mfa_last_step=-1,version=version+1 WHERE id=$1`, u.ID); e != nil {
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
		fmt.Println("MFA redefinido. Todas as sessões foram encerradas.")
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

func importLegacyAccounts(ctx context.Context, st *store.Store, cfg config.Config, assignments ...string) error {
	return importLegacyAccountsWithDefault(ctx, st, cfg, "", assignments...)
}

func importLegacyAccountsWithDefault(ctx context.Context, st *store.Store, cfg config.Config, fallback string, assignments ...string) error {
	if fallback != "" {
		parsed, e := uuid.Parse(fallback)
		if e != nil || parsed == uuid.Nil {
			return fmt.Errorf("--assign-unassigned-to exige um UUID de empresa válido")
		}
		fallback = parsed.String()
	}
	assigned, err := parseOrganizationAssignments(assignments)
	if err != nil {
		return err
	}
	used := map[string]bool{}
	kc := keycloak.NewAdminClient(cfg)
	if !kc.Enabled() {
		return fmt.Errorf("configure temporariamente KEYCLOAK_AUTH_SERVER_URL, KEYCLOAK_REALM e KEYCLOAK_ADMIN_PASSWORD para a migração")
	}
	users := []store.LocalUser{}
	for first := 0; ; first += 100 {
		page, e := kc.ListUsersPage(ctx, "", first, 100)
		if e != nil {
			return e
		}
		for _, old := range page {
			effective, e := kc.MigrationRoles(ctx, old.ID, cfg.KeycloakResource)
			if e != nil {
				return e
			}
			old.Roles = effective
			var org *auth.Organization
			if attr := old.Attributes["organization"]; len(attr) > 0 {
				org, _ = auth.ParseOrganizationClaim(attr[0])
				org = auth.NormalizeOrganization(org)
			}
			if org == nil || org.ID == "" {
				if ids := old.Attributes["organization_id"]; len(ids) > 0 {
					org = &auth.Organization{ID: ids[0]}
				}
			}
			if org == nil || org.ID == "" {
				if id, ok := assigned[old.Username]; ok {
					org = &auth.Organization{ID: id}
					used[old.Username] = true
				} else if fallback != "" {
					org = &auth.Organization{ID: fallback}
				} else {
					return fmt.Errorf("usuário %s sem organização: informe --assign-org %s=UUID_DA_EMPRESA para esta conta", old.Username, old.Username)
				}
			}

			if id, ok := assigned[old.Username]; ok && !used[old.Username] {
				if org.ID != id {
					return fmt.Errorf("usuário %s já pertence a outra organização; --assign-org não altera vínculos existentes", old.Username)
				}
				used[old.Username] = true
			}
			if _, e := uuid.Parse(org.ID); e != nil {
				return fmt.Errorf("organização inválida para %s", old.Username)
			}
			roles := []string{"user"}
			for _, role := range old.Roles {
				switch role {
				case "master", "admin", "financial", "sales", "employee", "operator", "viewer", "partner":
					roles = append(roles, role)
				}
			}
			doc := ""
			for _, key := range []string{"cpf_cnpj", "document", "cpf", "cnpj"} {
				if v := old.Attributes[key]; len(v) > 0 {
					doc = v[0]
					break
				}
			}
			users = append(users, store.LocalUser{ID: old.ID, Username: old.Username, Email: old.Email, FirstName: old.FirstName, LastName: old.LastName, OrganizationID: org.ID, OrganizationName: org.Name, Roles: roles, Document: doc, Enabled: old.Enabled})
		}
		if len(page) < 100 {
			break
		}
	}
	for name := range assigned {
		if !used[name] {
			return fmt.Errorf("usuário %s informado em --assign-org não encontrado; nenhuma conta importada", name)
		}
	}
	// Atomic and repeatable; existing local credentials are never overwritten.
	tx, e := st.Pool().Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	count := int64(0)
	for _, u := range users {
		tag, err := tx.Exec(ctx, `INSERT INTO auth_users(id,username,email,first_name,last_name,organization_id,organization_name,roles,document,enabled)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)ON CONFLICT(id)DO NOTHING`, u.ID, u.Username, u.Email, u.FirstName, u.LastName, u.OrganizationID, u.OrganizationName, u.Roles, u.Document, u.Enabled)
		if err != nil {
			return err
		}
		count += tag.RowsAffected()
	}
	if e = tx.Commit(ctx); e != nil {
		return e
	}
	fmt.Printf("%d contas importadas, IDs e empresas preservados. Defina novas senhas com auth reset-password.\n", count)
	return nil
}
