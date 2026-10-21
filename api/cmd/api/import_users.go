package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

// Account imports are offline and administrative. Credentials and MFA secrets
// from another provider are deliberately never imported.
type importedAccount struct {
	ID               string              `json:"id"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	Enabled          bool                `json:"enabled"`
	OrganizationID   string              `json:"organizationId"`
	OrganizationName string              `json:"organizationName"`
	Document         string              `json:"document"`
	Attributes       map[string][]string `json:"attributes"`
	Roles            []string            `json:"roles"`
	RealmRoles       []string            `json:"realmRoles"`
	ClientRoles      map[string][]string `json:"clientRoles"`
}

func importAccountsFile(ctx context.Context, st *store.Store, filename, fallback string, assignments ...string) error {
	var input io.Reader = os.Stdin
	if filename != "-" {
		file, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("abrir arquivo de contas: %w", err)
		}
		defer file.Close()
		input = file
	}
	users, err := readImportedAccounts(input, fallback, assignments...)
	if err != nil {
		return err
	}
	// Validate the whole file before writing. Repeat imports preserve local
	// identities, credentials, tenant bindings and MFA enrollment.
	tx, err := st.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	count := int64(0)
	for _, user := range users {
		tag, err := tx.Exec(ctx, `INSERT INTO auth_users(id,username,email,first_name,last_name,organization_id,organization_name,roles,document,enabled)VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)ON CONFLICT(id)DO NOTHING`, user.ID, user.Username, user.Email, user.FirstName, user.LastName, user.OrganizationID, user.OrganizationName, user.Roles, user.Document, user.Enabled)
		if err != nil {
			return fmt.Errorf("importar conta %s: %w", user.Username, err)
		}
		count += tag.RowsAffected()
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	fmt.Printf("%d contas importadas, IDs e empresas preservados. Defina novas senhas com auth reset-password.\n", count)
	return nil
}

func readImportedAccounts(input io.Reader, fallback string, assignments ...string) ([]store.LocalUser, error) {
	if fallback != "" {
		parsed, err := uuid.Parse(fallback)
		if err != nil || parsed == uuid.Nil {
			return nil, fmt.Errorf("--assign-unassigned-to exige um UUID de empresa válido")
		}
		fallback = parsed.String()
	}
	assigned, err := parseOrganizationAssignments(assignments)
	if err != nil {
		return nil, err
	}
	const maxBytes = 32 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(input, maxBytes+1))
	if err != nil {
		return nil, err
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || len(data) > maxBytes {
		return nil, fmt.Errorf("arquivo de contas vazio ou maior que 32 MiB")
	}
	var records []importedAccount
	if data[0] == '[' {
		err = json.Unmarshal(data, &records)
	} else {
		var export struct {
			Users []importedAccount `json:"users"`
		}
		err = json.Unmarshal(data, &export)
		records = export.Users
	}
	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("JSON inválido: informe uma lista de contas ou um objeto com users")
	}
	users := make([]store.LocalUser, 0, len(records))
	used, ids, names := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, old := range records {
		id, err := uuid.Parse(old.ID)
		username := strings.TrimSpace(old.Username)
		if err != nil || id == uuid.Nil || username == "" || len(username) > 128 || strings.ContainsAny(username, " \t\r\n") {
			return nil, fmt.Errorf("conta inválida: ID UUID e usuário sem espaços são obrigatórios")
		}
		nameKey := strings.ToLower(username)
		if ids[id.String()] || names[nameKey] {
			return nil, fmt.Errorf("ID ou usuário duplicado no arquivo: %s", username)
		}
		ids[id.String()], names[nameKey] = true, true
		org := &auth.Organization{ID: strings.TrimSpace(old.OrganizationID), Name: strings.TrimSpace(old.OrganizationName)}
		if org.ID == "" {
			if values := old.Attributes["organization"]; len(values) > 0 {
				var organizations map[string]json.RawMessage
				if len(values) != 1 || json.Unmarshal([]byte(values[0]), &organizations) != nil || len(organizations) != 1 {
					return nil, fmt.Errorf("organização ambígua para %s: informe organizationId explícito no arquivo", username)
				}
				parsed, e := auth.ParseOrganizationClaim(values[0])
				if e != nil {
					return nil, fmt.Errorf("organização inválida para %s", username)
				}
				if parsed != nil {
					org = auth.NormalizeOrganization(parsed)
				}
			}
		}
		if org.ID == "" {
			if values := old.Attributes["organization_id"]; len(values) > 0 {
				org.ID = strings.TrimSpace(values[0])
			}
		}
		if explicit, ok := assigned[username]; ok {
			if org.ID != "" && org.ID != explicit {
				return nil, fmt.Errorf("usuário %s já pertence a outra organização; --assign-org não altera vínculos existentes", username)
			}
			org.ID, used[username] = explicit, true
		}
		if org.ID == "" {
			org.ID = fallback
		}
		orgID, err := uuid.Parse(org.ID)
		if err != nil || orgID == uuid.Nil {
			return nil, fmt.Errorf("usuário %s sem organização válida: informe --assign-org %s=UUID_DA_EMPRESA", username, username)
		}
		roles, seen := []string{"user"}, map[string]bool{"user": true}
		// Only roles of this application are imported; foreign client roles
		// cannot grant platform privileges.
		allRoles := append(append(append([]string{}, old.Roles...), old.RealmRoles...), old.ClientRoles["connect-cli"]...)
		for _, role := range allRoles {
			switch role {
			case "master", "admin", "financial", "sales", "employee", "operator", "viewer", "partner":
				if !seen[role] {
					roles, seen[role] = append(roles, role), true
				}
			}
		}
		document := old.Document
		if document == "" {
			for _, key := range []string{"cpf_cnpj", "document", "cpf", "cnpj"} {
				if values := old.Attributes[key]; len(values) > 0 {
					document = values[0]
					break
				}
			}
		}
		users = append(users, store.LocalUser{ID: id.String(), Username: username, Email: old.Email, FirstName: old.FirstName, LastName: old.LastName, OrganizationID: orgID.String(), OrganizationName: org.Name, Roles: roles, Document: document, Enabled: old.Enabled})
	}
	for name := range assigned {
		if !used[name] {
			return nil, fmt.Errorf("usuário %s informado em --assign-org não encontrado; nenhuma conta importada", name)
		}
	}
	return users, nil
}
