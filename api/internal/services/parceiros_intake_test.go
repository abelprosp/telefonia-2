package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/parceiros"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

type parceirosTestStorage struct {
	uploads []string
}

func (s *parceirosTestStorage) PutIntakeObject(_ context.Context, bucket, key, _ string, _ []byte) error {
	s.uploads = append(s.uploads, bucket+"/"+key)
	return nil
}

func TestImportParceirosSale(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("defina TEST_DATABASE_URL para testar a entrada no Postgres")
	}
	ctx := context.Background()
	st, err := store.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)

	storage := &parceirosTestStorage{}
	svc := &Service{Store: st, IntakeStorage: storage, IntakeBucket: "files"}
	orgID := uuid.NewString()
	suffix := fmt.Sprintf("%08d", time.Now().UnixNano()%100000000)
	cpf := "111" + suffix
	otherCPF := "222" + suffix
	line := "51" + suffix
	saleID := "sale-" + suffix

	first, err := svc.ImportParceirosSale(ctx, orgID, parceiros.SaleIntake{
		Sale: parceiros.Sale{
			ID: saleID, Protocol: "VND-" + suffix, LineNumber: line, SimType: "ESIM",
			OperatorName: "Vivo", PlanName: "Controle",
		},
		Client: parceiros.Client{
			Name: "Amanda Martins", Document: cpf, Phone: "sem telefone",
			Email: "nao-e-email", Address: "Mauro Vieira 500",
		},
		Attachments: []parceiros.Attachment{
			{ID: "doc-1", Type: "CONTRACT", Name: "contrato.pdf", MimeType: "application/pdf", ContentBase64: base64.StdEncoding.EncodeToString([]byte("contrato"))},
			{ID: "doc-2", Type: "SELFIE", Name: "ruim.jpg", MimeType: "image/jpeg", ContentBase64: "%%%"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !first.CustomerCreated || !first.PhoneLineCreated || first.AttachmentsAccepted != 1 || first.AttachmentsSkipped != 1 {
		t.Fatalf("primeira entrada: %+v", first)
	}
	if first.PhoneLineID == nil {
		t.Fatal("linha")
	}

	var origin, rawAddress string
	var review bool
	if err = st.Pool().QueryRow(ctx, `SELECT "Origin", COALESCE("RawAddress", ''), "ReviewPending" FROM "Customers" WHERE "Id" = $1`, first.CustomerID).
		Scan(&origin, &rawAddress, &review); err != nil {
		t.Fatal(err)
	}
	if origin != "luxus_parceiros" || rawAddress != "Mauro Vieira 500" || review {
		t.Fatalf("cliente origin %q address %q review %v", origin, rawAddress, review)
	}
	var addresses int
	if err = st.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM "CustomerAddresses" WHERE "CustomerId" = $1`, first.CustomerID).Scan(&addresses); err != nil {
		t.Fatal(err)
	}
	if addresses != 0 {
		t.Fatalf("endereço livre virou ficha estruturada: %d", addresses)
	}
	var sim string
	if err = st.Pool().QueryRow(ctx, `SELECT "SimType" FROM "PhoneLines" WHERE "Id" = $1`, *first.PhoneLineID).Scan(&sim); err != nil {
		t.Fatal(err)
	}
	if sim != "ESIM" {
		t.Fatalf("sim %q", sim)
	}
	if len(storage.uploads) != 1 || !strings.HasPrefix(storage.uploads[0], "files/"+orgID+"/parceiros/"+saleID+"/") {
		t.Fatalf("uploads %v", storage.uploads)
	}

	replay, err := svc.ImportParceirosSale(ctx, orgID, parceiros.SaleIntake{
		Sale:   parceiros.Sale{ID: saleID, Protocol: "VND-" + suffix},
		Client: parceiros.Client{Name: "Outro Nome", Document: cpf},
	})
	if err != nil {
		t.Fatal(err)
	}
	if replay.CustomerID != first.CustomerID || !replay.CustomerCreated || replay.AttachmentsAccepted != 1 {
		t.Fatalf("reenvio: %+v", replay)
	}
	var customers int
	if err = st.Pool().QueryRow(ctx, `SELECT COUNT(*) FROM "Customers" WHERE "OrganizationId" = $1`, orgID).Scan(&customers); err != nil {
		t.Fatal(err)
	}
	if customers != 1 {
		t.Fatalf("clientes %d", customers)
	}

	second, err := svc.ImportParceirosSale(ctx, orgID, parceiros.SaleIntake{
		Sale:   parceiros.Sale{ID: "other-" + suffix, Protocol: "VND-B-" + suffix},
		Client: parceiros.Client{Name: "Amanda de novo", Document: cpf},
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.CustomerCreated || second.CustomerID != first.CustomerID {
		t.Fatalf("mesmo documento: %+v", second)
	}
	kept := false
	for _, warning := range second.Warnings {
		if strings.Contains(warning, "ficha manual foi mantida") {
			kept = true
		}
	}
	if !kept {
		t.Fatalf("warnings %v", second.Warnings)
	}

	conflict, err := svc.ImportParceirosSale(ctx, orgID, parceiros.SaleIntake{
		Sale:   parceiros.Sale{ID: "conflict-" + suffix, Protocol: "VND-C-" + suffix, LineNumber: line},
		Client: parceiros.Client{Name: "Outra Pessoa", Document: otherCPF},
	})
	if err != nil {
		t.Fatal(err)
	}
	if conflict.CustomerID == first.CustomerID || conflict.PhoneLineCreated {
		t.Fatalf("conflito de linha: %+v", conflict)
	}
	var owner string
	if err = st.Pool().QueryRow(ctx, `
		SELECT "CustomerId" FROM "PhoneLineCustomerLinks"
		WHERE "PhoneLineId" = $1 AND "EndDate" IS NULL`, *first.PhoneLineID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if owner != first.CustomerID {
		t.Fatalf("vínculo mudou para %s", owner)
	}

	pending, err := svc.ImportParceirosSale(ctx, orgID, parceiros.SaleIntake{
		Sale:   parceiros.Sale{ID: "pending-" + suffix, Protocol: "VND-D-" + suffix},
		Client: parceiros.Client{Name: "Sem CPF", Document: "123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = st.Pool().QueryRow(ctx, `SELECT "ReviewPending" FROM "Customers" WHERE "Id" = $1`, pending.CustomerID).Scan(&review); err != nil {
		t.Fatal(err)
	}
	if !pending.CustomerCreated || !review {
		t.Fatalf("pendente %+v review %v", pending, review)
	}
}
