// Package axon provides extensible import scaffolding for Axon line spreadsheets.
//
// IMPORTANT: the definitive column layout must come from a real Axon export.
// This package does not invent headers, ICCID/EID mappings or status vocabularies.
package axon

import (
	"fmt"
	"strings"
)

// Layout describes a registered column mapping for a specific Axon file version.
type Layout struct {
	Code        string
	Description string
	// RequiredHeaders must match the spreadsheet header row exactly (case-insensitive trim).
	RequiredHeaders []string
	// Column keys are logical fields; values are header names from the real file.
	Columns map[string]string
}

var registry = map[string]Layout{}

// RegisterLayout adds a layout once the real Axon file headers are known.
func RegisterLayout(l Layout) error {
	code := strings.TrimSpace(strings.ToLower(l.Code))
	if code == "" {
		return fmt.Errorf("layout code required")
	}
	if len(l.RequiredHeaders) == 0 {
		return fmt.Errorf("layout %s: required headers must be provided from a real sample file", code)
	}
	l.Code = code
	registry[code] = l
	return nil
}

// GetLayout returns a registered layout or an error instructing operators to supply the real file.
func GetLayout(code string) (Layout, error) {
	code = strings.TrimSpace(strings.ToLower(code))
	if l, ok := registry[code]; ok {
		return l, nil
	}
	return Layout{}, fmt.Errorf(
		"layout %q não registrado: forneça um arquivo Axon real e registre o mapeamento de colunas antes de processar",
		code,
	)
}

// ValidateHeaders ensures the spreadsheet headers cover the layout's required set.
func ValidateHeaders(layout Layout, headers []string) error {
	norm := map[string]struct{}{}
	for _, h := range headers {
		norm[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
	}
	var missing []string
	for _, req := range layout.RequiredHeaders {
		if _, ok := norm[strings.ToLower(strings.TrimSpace(req))]; !ok {
			missing = append(missing, req)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("cabeçalhos ausentes no arquivo Axon: %s", strings.Join(missing, ", "))
	}
	return nil
}

// LogicalFields lists fields the importer can map when present in Columns.
var LogicalFields = []string{
	"phone_number", "status", "line_type", "iccid", "eid", "external_id", "customer_ref", "operator",
}
