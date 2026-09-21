package storage

import (
	"bytes"
	"io"
	"testing"
)

func TestMaxObjectBytesConstant(t *testing.T) {
	if MaxObjectBytes != DefaultMaxObjectBytes {
		t.Fatalf("unexpected max %d", MaxObjectBytes)
	}
	if DefaultMaxObjectBytes != 512*1024*1024 {
		t.Fatalf("expected 512MiB default, got %d", DefaultMaxObjectBytes)
	}
	if DefaultMaxDiskObjectBytes != 2*1024*1024*1024 {
		t.Fatalf("expected 2GiB disk default, got %d", DefaultMaxDiskObjectBytes)
	}
	if MaxDiskObjectBytes < MaxObjectBytes {
		t.Fatalf("disk ceiling %d should be >= memory ceiling %d", MaxDiskObjectBytes, MaxObjectBytes)
	}
}

func TestLimitReaderRejectsOversize(t *testing.T) {
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

func TestConfigureMaxObjectBytes(t *testing.T) {
	prev := MaxObjectBytes
	t.Cleanup(func() { MaxObjectBytes = prev })
	ConfigureMaxObjectBytes(100)
	if MaxObjectBytes != 100 {
		t.Fatalf("got %d", MaxObjectBytes)
	}
	ConfigureMaxObjectBytes(0)
	if MaxObjectBytes != 100 {
		t.Fatal("zero must not reset")
	}
}

func TestConfigureMaxDiskObjectBytes(t *testing.T) {
	prev := MaxDiskObjectBytes
	t.Cleanup(func() { MaxDiskObjectBytes = prev })
	ConfigureMaxDiskObjectBytes(2048)
	if MaxDiskObjectBytes != 2048 {
		t.Fatalf("got %d", MaxDiskObjectBytes)
	}
	ConfigureMaxDiskObjectBytes(0)
	if MaxDiskObjectBytes != 2048 {
		t.Fatal("zero must not reset")
	}
}
