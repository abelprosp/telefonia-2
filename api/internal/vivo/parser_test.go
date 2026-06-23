package vivo

import (
	"testing"
	"time"
)

func TestReadRecordType(t *testing.T) {
	line := stringsRepeat(' ', 110) + "010D" + stringsRepeat(' ', 50)
	got := ReadRecordType(line)
	if got != "010D" {
		t.Fatalf("ReadRecordType() = %q, want 010D", got)
	}
}

func TestParse010DHeader(t *testing.T) {
	line := buildLine(map[int]string{
		0:    "1234567890",
		61:   "01",
		84:   "ABC",
		110:  "010D",
		205:  "202406",
		229:  "20240615",
		247:  "20240625",
		341:  "000000100.50",
		353:  "20240601",
		361:  "20240630",
		372:  "000000020.00",
		540:  "000000120.50",
		1770: "FISC123456789",
	})

	records := Parse(line)
	if len(records) != 1 {
		t.Fatalf("Parse() len = %d, want 1", len(records))
	}

	header, ok := records[0].(*Line010DHeader)
	if !ok {
		t.Fatalf("record type = %T, want *Line010DHeader", records[0])
	}

	if header.AccountNumber != "1234567890" {
		t.Errorf("AccountNumber = %q", header.AccountNumber)
	}
	if header.ReferenceMonth != "202406" {
		t.Errorf("ReferenceMonth = %q", header.ReferenceMonth)
	}
	if header.TotalAmount != 120.50 {
		t.Errorf("TotalAmount = %v", header.TotalAmount)
	}
	if header.IssueDate.Format("20060102") != "20240615" {
		t.Errorf("IssueDate = %v", header.IssueDate)
	}
}

func TestParseSkipsUnknownRecordType(t *testing.T) {
	line := buildLine(map[int]string{
		110: "999Z",
	})
	records := Parse(line)
	if len(records) != 0 {
		t.Fatalf("Parse() len = %d, want 0 for unknown type", len(records))
	}
}

func TestFilterByType(t *testing.T) {
	line1 := buildLine(map[int]string{110: "010D", 0: "1111111111"})
	line2 := buildLine(map[int]string{110: "011D", 0: "2222222222"})
	records := Parse(line1 + "\n" + line2)

	headers := FilterByType[*Line010DHeader](records)
	if len(headers) != 1 {
		t.Fatalf("FilterByType headers len = %d, want 1", len(headers))
	}
	if headers[0].AccountNumber != "1111111111" {
		t.Errorf("header AccountNumber = %q", headers[0].AccountNumber)
	}
}

func TestDecodeLatin1(t *testing.T) {
	raw := []byte{0xE7, 0xE3, 0x6F} // "ção" partial: çã + o
	got := DecodeLatin1BestEffort(raw)
	if got == "" {
		t.Fatal("DecodeLatin1BestEffort returned empty")
	}
}

func buildLine(fields map[int]string) string {
	maxEnd := 0
	for offset, value := range fields {
		end := offset + len(value)
		if end > maxEnd {
			maxEnd = end
		}
	}
	if maxEnd < 120 {
		maxEnd = 120
	}
	buf := make([]byte, maxEnd)
	for i := range buf {
		buf[i] = ' '
	}
	for offset, value := range fields {
		copy(buf[offset:], value)
	}
	return string(buf)
}

func stringsRepeat(ch byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}

func TestParseDateFormat(t *testing.T) {
	d := bindDate("20240615", 0, 8, "yyyyMMdd", true)
	if d.IsZero() {
		t.Fatal("expected non-zero date")
	}
	want := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !d.Equal(want) {
		t.Errorf("date = %v, want %v", d, want)
	}
}
