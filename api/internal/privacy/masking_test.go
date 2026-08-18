package privacy

import "testing"

func TestMaskDocument_CPF(t *testing.T) {
	cpf := "123.456.789-01"
	masked := MaskDocument(cpf)
	expected := "123.***.***-01"
	if masked != expected {
		t.Errorf("MaskDocument(%q) = %q; want %q", cpf, masked, expected)
	}
}

func TestMaskDocument_CNPJ(t *testing.T) {
	cnpj := "12.345.678/0001-90"
	masked := MaskDocument(cnpj)
	expected := "12.***.***/0001-90"
	if masked != expected {
		t.Errorf("MaskDocument(%q) = %q; want %q", cnpj, masked, expected)
	}
}

func TestMaskEmail(t *testing.T) {
	email := "joao.silva@empresa.com.br"
	masked := MaskEmail(email)
	expected := "jo********@empresa.com.br"
	if masked != expected {
		t.Errorf("MaskEmail(%q) = %q; want %q", email, masked, expected)
	}
}

func TestMaskPhone(t *testing.T) {
	phone := "11987654321"
	masked := MaskPhone(phone)
	expected := "(11) 9****-**21"
	if masked != expected {
		t.Errorf("MaskPhone(%q) = %q; want %q", phone, masked, expected)
	}
}
