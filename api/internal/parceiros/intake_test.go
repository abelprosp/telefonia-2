package parceiros

import "testing"

func TestPrepareAceitaVendaIncompletaDoParceiros(t *testing.T) {
	prepared, err := Prepare(SaleIntake{
		Sale: Sale{
			ID: "sale-1", Protocol: "VND-1", TaskProtocol: "LUX-2026-1",
			LineNumber: "51 98537-0873", SimType: "ESIM",
			DeviceImei: "123", DeviceEid: "999", DevicePlatform: "ios",
			SoldAt: "2026-05-26T14:55:22.877Z",
		},
		Client: Client{
			Name: "Amanda Martins", Document: "021.134.190-84",
			Phone: "sem telefone", Email: "nao-e-email",
			Address: "Mauro Vieira 500",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !prepared.LookupDocument || prepared.DocumentDigits != "02113419084" || prepared.CustomerType != "pf" {
		t.Fatalf("documento = %+v", prepared)
	}
	if prepared.Phone != "" {
		t.Fatalf("telefone %q", prepared.Phone)
	}
	if prepared.Email != "" || prepared.RawEmail != "nao-e-email" {
		t.Fatalf("email %q raw %q", prepared.Email, prepared.RawEmail)
	}
	if prepared.StructuredAddress != nil || prepared.RawAddress != "Mauro Vieira 500" {
		t.Fatalf("endereco %+v raw %q", prepared.StructuredAddress, prepared.RawAddress)
	}
	if prepared.LineDigits != "51985370873" || prepared.SimType != "ESIM" {
		t.Fatalf("linha %q sim %q", prepared.LineDigits, prepared.SimType)
	}
	if prepared.SoldAt == nil {
		t.Fatal("data da venda")
	}
	if len(prepared.Warnings) < 3 {
		t.Fatalf("warnings %v", prepared.Warnings)
	}
}

func TestPrepareDocumentoInvalidoNaoBloqueia(t *testing.T) {
	prepared, err := Prepare(SaleIntake{
		Sale:   Sale{ID: "sale-2", Protocol: "VND-2"},
		Client: Client{Name: "Sem CPF", Document: "123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.LookupDocument || !prepared.ReviewPending || prepared.RawDocument != "123" {
		t.Fatalf("%+v", prepared)
	}
}

func TestPrepareEnderecoCompletoViraCampos(t *testing.T) {
	prepared, err := Prepare(SaleIntake{
		Sale: Sale{ID: "sale-3", Protocol: "VND-3", SimType: "CHIP"},
		Client: Client{
			Name: "Completo", Document: "12345678901",
			Address: "Rua A", AddressNumber: "10", Neighborhood: "Centro",
			City: "Lajeado", State: "rs", ZipCode: "95900-000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.StructuredAddress == nil || prepared.StructuredAddress.State != "RS" || prepared.StructuredAddress.ZipCode != "95900000" {
		t.Fatalf("%+v", prepared.StructuredAddress)
	}
	if prepared.SimType != "PHYSICAL" {
		t.Fatalf("sim %q", prepared.SimType)
	}
}

func TestPrepareExigeIdentidadeDaVenda(t *testing.T) {
	_, err := Prepare(SaleIntake{Client: Client{Name: "Amanda"}})
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestPrepareRejeitaIDComTravessiaDeCaminho(t *testing.T) {
	for _, id := range []string{"../escape", "..", "a/b", `a\b`, ".hidden", "sale\x00id"} {
		_, err := Prepare(SaleIntake{
			Sale:   Sale{ID: id, Protocol: "VND-1"},
			Client: Client{Name: "Amanda"},
		})
		if err == nil {
			t.Fatalf("sale.id %q deveria ser rejeitado", id)
		}
	}
}
