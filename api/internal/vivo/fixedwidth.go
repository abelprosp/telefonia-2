package vivo

import (
	"strconv"
	"strings"
	"time"
)

const (
	recordTypeOffset = 110
	recordTypeLength = 4
)

// ParserOptions controla leitura do tipo de registro e fatias de campo.
type ParserOptions struct {
	RecordTypeOffset  int
	RecordTypeLength  int
	TrimRecordType    bool
	TrimFieldValues   bool
}

// DefaultParserOptions espelha FixedWidthTextParserOptions do C#.
func DefaultParserOptions() ParserOptions {
	return ParserOptions{
		RecordTypeOffset: recordTypeOffset,
		RecordTypeLength: recordTypeLength,
		TrimRecordType:   true,
		TrimFieldValues:  true,
	}
}

func readRecordType(line string, opts ParserOptions) string {
	o := opts.RecordTypeOffset
	length := opts.RecordTypeLength
	if o < 0 || length <= 0 || o >= len(line) {
		return ""
	}
	take := length
	if o+take > len(line) {
		take = len(line) - o
	}
	s := line[o : o+take]
	if opts.TrimRecordType {
		s = strings.TrimSpace(s)
	}
	return s
}

func sliceField(line string, offset, length int, trim bool) string {
	if offset < 0 || length <= 0 {
		return ""
	}
	if offset >= len(line) {
		return ""
	}
	end := offset + length
	if end > len(line) {
		end = len(line)
	}
	s := line[offset:end]
	if trim {
		s = strings.TrimSpace(s)
	}
	return s
}

func bindString(line string, offset, length int, trim bool) string {
	return sliceField(line, offset, length, trim)
}

func bindInt(line string, offset, length int, trim bool) int {
	s := sliceField(line, offset, length, trim)
	if strings.TrimSpace(s) == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func bindDecimal(line string, offset, length int, trim bool) float64 {
	s := sliceField(line, offset, length, trim)
	if strings.TrimSpace(s) == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func bindDate(line string, offset, length int, format string, trim bool) time.Time {
	s := sliceField(line, offset, length, trim)
	if strings.TrimSpace(s) == "" {
		return time.Time{}
	}
	layout := normalizeDateFormat(format)
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func normalizeDateFormat(format string) string {
	switch strings.TrimSpace(format) {
	case "DDmmyyyy", "ddMMyyyy", "DDMMYYYY":
		return "02012006"
	case "yyyyMMdd", "YYYYMMDD":
		return "20060102"
	default:
		if format == "" {
			return "2006-01-02"
		}
		return format
	}
}

func splitLines(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	raw := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\r' || r == '\n'
	})
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
