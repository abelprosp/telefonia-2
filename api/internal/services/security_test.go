package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestContractHTMLRejectsStoredAndCustomerXSS(t *testing.T) {
	customer := &store.CustomerContractData{Name: `<img src=x onerror="alert(1)">`, Street: `"><svg onload="alert(2)">`}
	sale := &models.GetSaleResponse{}
	result := renderContractTemplate(`<p>{{customer.name}} {{customer.address.street}}</p><script>alert(3)</script><a href="javascript:alert(4)">link</a>`, customer, sale, nil)
	if strings.Contains(result, "<img") || strings.Contains(result, "<svg") || strings.Contains(result, "<script") || strings.Contains(result, "javascript:") {
		t.Fatalf("unsafe HTML: %s", result)
	}
	if !strings.Contains(result, "&lt;img") || !strings.Contains(result, "<p>") {
		t.Fatalf("customer text or safe markup lost: %s", result)
	}
}

type recordingStorage struct{ keys []string }

func (s *recordingStorage) CreatePresignedUploadURL(_ context.Context, _, key string, _ time.Duration, _ *string) (*models.PresignedURLModel, error) {
	s.keys = append(s.keys, key)
	return &models.PresignedURLModel{URL: "https://storage.example/upload"}, nil
}
func (s *recordingStorage) CreatePresignedDownloadURL(_ context.Context, _, key string, _ time.Duration) (*models.PresignedURLModel, error) {
	s.keys = append(s.keys, key)
	return &models.PresignedURLModel{URL: "https://storage.example/download"}, nil
}

func storageContext(role, org string) context.Context {
	ctx := auth.WithUser(context.Background(), &auth.User{ID: "user", Roles: []string{role}})
	return auth.WithOrganization(ctx, &auth.Organization{ID: org})
}

func TestGenericStorageRejectsPortalPartnerAndSales(t *testing.T) {
	storage := &recordingStorage{}
	svc := &PresignedService{Storage: storage, BucketName: "files"}
	for _, role := range []string{auth.RoleUser, auth.RolePartner, auth.RoleSales} {
		ctx := storageContext(role, "tenant")
		if _, err := svc.CreateDownloadURL(ctx, models.CreatePresignedDownloadURLInput{BucketName: "files", ObjectKey: "tenant/customer-B.pdf"}); err == nil {
			t.Fatalf("%s can download another customer's object", role)
		}
		if _, err := svc.CreateUploadURL(ctx, models.CreatePresignedUploadURLInput{BucketName: "files", ObjectKey: "existing.pdf"}); err == nil {
			t.Fatalf("%s can upload arbitrary objects", role)
		}
	}
	if len(storage.keys) != 0 {
		t.Fatal("unauthorized request reached storage")
	}
}

func TestUploadsCannotOverwriteCallerSelectedKey(t *testing.T) {
	storage := &recordingStorage{}
	svc := &PresignedService{Storage: storage, BucketName: "files"}
	ctx := storageContext(auth.RoleEmployee, "tenant")
	input := models.CreatePresignedUploadURLInput{BucketName: "files", ObjectKey: "tenant/existing.pdf"}
	a, err := svc.CreateUploadURL(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.CreateUploadURL(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if a.ObjectKey == input.ObjectKey || a.ObjectKey == b.ObjectKey || !strings.HasPrefix(a.ObjectKey, "tenant/uploads/") {
		t.Fatalf("unsafe upload keys: %s %s", a.ObjectKey, b.ObjectKey)
	}
	if _, err := svc.CreateDownloadURL(ctx, models.CreatePresignedDownloadURLInput{BucketName: "other", ObjectKey: a.ObjectKey}); err == nil {
		t.Fatal("foreign bucket accepted")
	}
}

func TestAuthorizedTicketStorageRejectsForeignTenantEvenForLegacyLuxus(t *testing.T) {
	svc := &PresignedService{Storage: &recordingStorage{}, BucketName: "files"}
	ctx := storageContext(auth.RoleUser, store.DefaultLuxusOrgID)
	if _, err := svc.CreateAuthorizedDownloadURL(ctx, models.CreatePresignedDownloadURLInput{BucketName: "files", ObjectKey: "tickets/other-tenant/ticket/file.pdf"}); err == nil {
		t.Fatal("foreign ticket accepted")
	}
	if _, err := svc.CreateAuthorizedDownloadURL(ctx, models.CreatePresignedDownloadURLInput{BucketName: "files", ObjectKey: "tickets/" + store.DefaultLuxusOrgID + "/ticket/file.pdf"}); err != nil {
		t.Fatal(err)
	}
}

func TestPasswordChangeCannotUseAdminReset(t *testing.T) {
	password := "new-password"
	svc := &Service{}
	if _, err := svc.UpdateCurrentUserProfile(context.Background(), models.UpdateUserProfileInput{NewPassword: &password}); err == nil {
		t.Fatal("password change must require the authenticated account security flow")
	}
}
