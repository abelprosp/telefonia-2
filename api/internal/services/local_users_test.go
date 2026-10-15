package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"os"
	"testing"
)

func TestLocalUserAdministrationCannotCrossTenantOrRemoveOwnAccess(t *testing.T) {
	database := os.Getenv("AUTH_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a disposable database")
	}
	base := context.Background()
	st, e := store.New(base, database)
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	org, other := uuid.NewString(), uuid.NewString()
	hash, _ := localauth.HashPassword("test safe password 2026")
	first := &store.LocalUser{ID: uuid.NewString(), Username: "admin-" + uuid.NewString(), Email: "admin@example.com", OrganizationID: org, Roles: []string{"master", "user"}, Enabled: true, PasswordHash: hash}
	second := &store.LocalUser{ID: uuid.NewString(), Username: "foreign-" + uuid.NewString(), Email: "foreign@example.com", OrganizationID: other, Roles: []string{"master", "user"}, Enabled: true, PasswordHash: hash}
	for _, u := range []*store.LocalUser{first, second} {
		if e = st.CreateLocalUser(base, u); e != nil {
			t.Fatal(e)
		}
		defer st.Pool().Exec(base, `DELETE FROM auth_users WHERE id=$1`, u.ID)
	}
	ctx := auth.WithUser(auth.WithOrganization(base, &auth.Organization{ID: org, Name: "Org A"}), &auth.User{ID: first.ID, Roles: first.Roles})
	svc := &Service{Store: st}
	list, e := svc.ListOrganizationUsers(ctx, "")
	if e != nil || len(list) != 1 || list[0].ID != first.ID {
		t.Fatal("list leaked foreign users", e)
	}
	enabled := false
	if _, e = svc.UpdateOrganizationUser(ctx, second.ID, models.UpdateOrganizationUserInput{Enabled: &enabled}); e == nil {
		t.Fatal("foreign user modified")
	}
	if _, e = svc.UpdateOrganizationUser(ctx, first.ID, models.UpdateOrganizationUserInput{Enabled: &enabled}); e == nil {
		t.Fatal("self lockout accepted")
	}
	employee := "employee"
	if _, e = svc.UpdateOrganizationUser(ctx, first.ID, models.UpdateOrganizationUserInput{Profile: &employee}); e == nil {
		t.Fatal("self demotion accepted")
	}
	created, e := svc.CreateOrganizationUser(ctx, models.CreateOrganizationUserInput{Username: "new-" + uuid.NewString(), Email: "new@example.com", Password: "new safe password 2026", Profile: "sales"})
	if e != nil {
		t.Fatal(e)
	}
	defer st.Pool().Exec(base, `DELETE FROM auth_users WHERE id=$1`, created.ID)
	saved, e := st.LocalUserByID(base, created.ID)
	if e != nil || saved.OrganizationID != org || !localauth.CheckPassword(saved.PasswordHash, "new safe password 2026") {
		t.Fatal("new local account invalid", e)
	}
	if len(saved.Roles) != 2 || saved.Roles[0] != "sales" {
		t.Fatal("new account received excessive permissions")
	}
}
