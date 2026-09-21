package importservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/storage"
	"github.com/luxus-connect/telefonia/api/internal/vivo"
)

// ObjectMaterializer streams invoice files to disk for large imports.
type ObjectMaterializer interface {
	MaterializeObject(ctx context.Context, bucket, key string) (*storage.MaterializedObject, error)
}

type loadInvoiceResult struct {
	Path   string
	SHA256 string
	Parsed []any
	IsPDF  bool
}

func (r *loadInvoiceResult) Cleanup() {
	if r != nil && r.Path != "" {
		_ = os.Remove(r.Path)
	}
}

func (p *Processor) loadInvoiceForProcessing(ctx context.Context, bucket, key string) (*loadInvoiceResult, error) {
	if p.Storage == nil {
		return nil, httputil.BusinessError(notifications.ObjectStorageUnavailable)
	}

	// Prefer disk-backed streaming path when available.
	if m, ok := p.Storage.(ObjectMaterializer); ok {
		mat, err := m.MaterializeObject(ctx, bucket, key)
		if err != nil {
			return nil, httputil.BusinessError(notifications.ImportFileReadError)
		}
		head, err := peekFileHead(mat.Path, 8)
		if err != nil {
			_ = os.Remove(mat.Path)
			return nil, httputil.BusinessError(notifications.ImportFileReadError)
		}
		if isPDFBytes(head) {
			_ = os.Remove(mat.Path)
			return &loadInvoiceResult{SHA256: mat.SHA256, IsPDF: true}, nil
		}
		parsed, err := vivo.ParseLatin1File(mat.Path)
		if err != nil {
			_ = os.Remove(mat.Path)
			return nil, httputil.BusinessError(notifications.N("IMPORT_PARSE_FAILED", "Não foi possível interpretar o TXT VIVO."))
		}
		return &loadInvoiceResult{Path: mat.Path, SHA256: mat.SHA256, Parsed: parsed}, nil
	}

	// Fallback: in-memory GetObject (tests / legacy).
	raw, err := p.Storage.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError("storage get: " + err.Error()))
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	if isPDFBytes(raw) {
		return &loadInvoiceResult{SHA256: hash, IsPDF: true}, nil
	}
	parsed, err := vivo.ParseLatin1(raw)
	if err != nil {
		return nil, httputil.BusinessError(notifications.N("IMPORT_PARSE_FAILED", "Não foi possível interpretar o TXT VIVO."))
	}
	return &loadInvoiceResult{SHA256: hash, Parsed: parsed}, nil
}

func peekFileHead(path string, n int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, n)
	read, err := f.Read(buf)
	if err != nil && read == 0 {
		return nil, err
	}
	return buf[:read], nil
}
