package vivo

import (
	"bufio"
	"io"
	"os"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// ParseLatin1Reader streams an ISO-8859-1 (or best-effort) invoice without loading
// the entire file into a single []byte first. Lines are decoded and deserialized
// incrementally.
func ParseLatin1Reader(r io.Reader) ([]any, error) {
	if r == nil {
		return nil, io.ErrUnexpectedEOF
	}
	decoded := transform.NewReader(r, charmap.ISO8859_1.NewDecoder())
	sc := bufio.NewScanner(decoded)
	// VIVO fixed-width lines can be long; allow up to 1 MiB per line.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)

	opts := DefaultParserOptions()
	out := make([]any, 0, 1024)
	lines := 0
	for sc.Scan() {
		line := sc.Text()
		lines++
		if rec := tryDeserializeLine(line, opts); rec != nil {
			out = append(out, rec)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if lines == 0 && len(out) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	return out, nil
}

// ParseLatin1File opens path and parses via streaming reader.
func ParseLatin1File(path string) ([]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseLatin1Reader(f)
}
