package httputil

import (
	"fmt"
	"strings"
	"time"
)

var dateOnlyLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02",
}

func parseDateString(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("invalid date format")
	}
	for _, layout := range dateOnlyLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

// ParseOptionalDate accepts RFC3339 or yyyy-MM-dd date strings.
func ParseOptionalDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	value := strings.TrimSpace(*raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseDateString(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
