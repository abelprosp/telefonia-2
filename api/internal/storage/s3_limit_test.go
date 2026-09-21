package storage

import (
	"bytes"
	"io"
	"testing"
)

func TestMaxObjectBytesConstant(t *testing.T) {
	if MaxObjectBytes != 256*1024*1024 {
		t.Fatalf("unexpected max %d", MaxObjectBytes)
	}
}

func TestLimitReaderDetectsOversizedPayload(t *testing.T) {
	big := bytes.Repeat([]byte("a"), int(MaxObjectBytes)+10)
	limited := io.LimitReader(bytes.NewReader(big), MaxObjectBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(buf)) <= MaxObjectBytes {
		t.Fatal("expected oversize buffer from LimitReader+1")
	}
}
