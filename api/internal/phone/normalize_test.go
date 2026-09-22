package phone

import "testing"

func TestNormalize_masksAndSpaces(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"(51) 99999-8888", "51999998888"},
		{"51 99999 8888", "51999998888"},
		{"+55 (51) 99999-8888", "5551999998888"},
		{"51999998888", "51999998888"},
		{"  (11) 3456-7890  ", "1134567890"},
		{"", ""},
		{"   ", ""},
		{"abc", ""},
		{"51-ABCD-9999", "519999"},
	}
	for _, tc := range cases {
		got := Normalize(tc.in)
		if got.Normalized != tc.want {
			t.Errorf("Normalize(%q).Normalized = %q, want %q", tc.in, got.Normalized, tc.want)
		}
		if tc.in != "" && stringsTrim(tc.in) != "" && got.Original != stringsTrim(tc.in) {
			t.Errorf("Normalize(%q).Original = %q, want trimmed input", tc.in, got.Original)
		}
	}
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func TestNormalize_doesNotStripDDI55(t *testing.T) {
	got := Digits("+55 51 99999-8888")
	if got != "5551999998888" {
		t.Fatalf("DDI 55 must be preserved, got %q", got)
	}
}

func TestEqual(t *testing.T) {
	if !Equal("(51) 99999-8888", "51999998888") {
		t.Fatal("expected equal after normalize")
	}
	if Equal("", "") {
		t.Fatal("empty must not equal")
	}
	if Equal("51999998888", "51999998889") {
		t.Fatal("different numbers must not equal")
	}
}

func equal(a, b string) bool { return Equal(a, b) }

func TestIsValidBasic(t *testing.T) {
	if IsValidBasic("", 10, 13) {
		t.Fatal("empty invalid")
	}
	if IsValidBasic("123", 10, 13) {
		t.Fatal("too short")
	}
	if !IsValidBasic("(51) 99999-8888", 10, 13) {
		t.Fatal("valid BR mobile")
	}
	if !IsValidBasic("+55 51 99999-8888", 10, 13) {
		t.Fatal("with DDI still within 13")
	}
	if IsValidBasic("123456789012345", 10, 13) {
		t.Fatal("too long")
	}
}

func TestDigits_duplicatesFormats(t *testing.T) {
	a := Digits("(51) 9.9999-8888")
	b := Digits("51 99999 8888")
	if a != b || a != "51999998888" {
		t.Fatalf("got %q and %q", a, b)
	}
}

func TestNormalizeSimType(t *testing.T) {
	cases := map[string]string{
		"":           SimUnknown,
		"physical":   SimPhysical,
		"SIM":        SimPhysical,
		"e-sim":      SimEsim,
		"digital":    SimEsim,
		"unknown":    SimUnknown,
		"chip":       SimPhysical,
		"weird":      SimUnknown,
	}
	for in, want := range cases {
		if got := NormalizeSimType(in); got != want {
			t.Errorf("NormalizeSimType(%q)=%q want %q", in, got, want)
		}
	}
}
