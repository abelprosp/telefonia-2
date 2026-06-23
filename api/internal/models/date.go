package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var dateOnlyLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02",
}

// DateInput accepts RFC3339 or yyyy-MM-dd in JSON payloads.
type DateInput struct {
	time.Time
}

func (t *DateInput) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	parsed, err := parseDateString(raw)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
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
