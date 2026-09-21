package vivo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLatin1Reader_emptyEOF(t *testing.T) {
	_, err := ParseLatin1Reader(strings.NewReader(""))
	if err == nil {
		t.Fatal("expected EOF on empty")
	}
}

func TestParseLatin1Reader_lineByLine(t *testing.T) {
	// Unknown record types are skipped; must not panic / EOF on multi-line input.
	payload := strings.Repeat("X", 120) + "010D" + strings.Repeat(" ", 200) + "\n"
	payload += strings.Repeat("Y", 120) + "011D" + strings.Repeat(" ", 200) + "\n"
	recs, err := ParseLatin1Reader(strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	_ = recs
}

func TestParseLatin1File_readsFromDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invoice.txt")
	payload := strings.Repeat("Z", 80) + "\n" + strings.Repeat("W", 80) + "\n"
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	recs, err := ParseLatin1File(path)
	if err != nil {
		t.Fatal(err)
	}
	_ = recs
}
