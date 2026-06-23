package vivo

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const ProviderName = "VIVO"

var registry = buildDeserializerRegistry()

// Parse deserializa texto já decodificado (equivalente a VivoTextInvoiceParser.Parse no C#).
// Retorna uma slice de any com ponteiros para os tipos de linha correspondentes.
func Parse(text string) []any {
	return parseWithOptions(text, DefaultParserOptions())
}

// ParseWithOptions deserializa com opções customizadas de largura fixa.
func ParseWithOptions(text string, opts ParserOptions) []any {
	return parseWithOptions(text, opts)
}

func parseWithOptions(text string, opts ParserOptions) []any {
	lines := splitLines(text)
	out := make([]any, 0, len(lines))
	for _, line := range lines {
		if rec := tryDeserializeLine(line, opts); rec != nil {
			out = append(out, rec)
		}
	}
	return out
}

// ParseLatin1 decodifica bytes ISO-8859-1 e deserializa o conteúdo.
func ParseLatin1(raw []byte) ([]any, error) {
	if len(raw) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	text := DecodeLatin1BestEffort(raw)
	return Parse(text), nil
}

// DecodeLatin1BestEffort decodifica ISO-8859-1; em falha usa UTF-8 (como ProcessImportInvoiceCommandHandler).
func DecodeLatin1BestEffort(raw []byte) string {
	decoded, err := charmap.ISO8859_1.NewDecoder().Bytes(raw)
	if err != nil {
		return string(raw)
	}
	return string(decoded)
}

// DecodeLatin1 decodifica estritamente ISO-8859-1.
func DecodeLatin1(raw []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(raw), charmap.ISO8859_1.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// ReadRecordType lê o código de registro na posição configurada (offset 110, largura 4).
func ReadRecordType(line string) string {
	return readRecordType(line, DefaultParserOptions())
}

// TryDeserializeLine tenta deserializar uma única linha.
func TryDeserializeLine(line string) (any, bool) {
	rec := tryDeserializeLine(line, DefaultParserOptions())
	return rec, rec != nil
}

func tryDeserializeLine(line string, opts ParserOptions) any {
	if strings.TrimSpace(line) == "" {
		return nil
	}
	rt := readRecordType(line, opts)
	if rt == "" {
		return nil
	}
	deserialize, ok := registry[rt]
	if !ok {
		return nil
	}
	return deserialize(line, opts.TrimFieldValues)
}

// RegisteredRecordTypes retorna os tipos de registro suportados.
func RegisteredRecordTypes() []string {
	types := make([]string, 0, len(registry))
	for rt := range registry {
		types = append(types, rt)
	}
	return types
}

// FilterByType retorna registros do tipo T da slice parseada.
func FilterByType[T any](records []any) []T {
	out := make([]T, 0)
	for _, rec := range records {
		if typed, ok := rec.(T); ok {
			out = append(out, typed)
		}
	}
	return out
}

// FirstByType retorna o primeiro registro do tipo T, se existir.
func FirstByType[T any](records []any) (T, bool) {
	var zero T
	for _, rec := range records {
		if typed, ok := rec.(T); ok {
			return typed, true
		}
	}
	return zero, false
}
