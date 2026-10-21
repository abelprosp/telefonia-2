package localauth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Fixed, bounded parameters: 19 MiB, two iterations, one lane (OWASP minimum).
func HashPassword(password string) (string, error) {
	if len([]rune(password)) < 12 || len(password) > 256 {
		return "", fmt.Errorf("a senha deve ter entre 12 caracteres e 256 bytes")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32)
	return "$argon2id$v=19$m=19456,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}

func CheckPassword(encoded, password string) bool {
	if len(password) > 256 {
		return false
	}
	p := strings.Split(encoded, "$")
	if len(p) != 6 || p[1] != "argon2id" || p[2] != "v=19" || p[3] != "m=19456,t=2,p=1" {
		return false
	}
	salt, e1 := base64.RawStdEncoding.DecodeString(p[4])
	want, e2 := base64.RawStdEncoding.DecodeString(p[5])
	if e1 != nil || e2 != nil || len(salt) != 16 || len(want) != 32 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, 2, 19*1024, 1, 32)
	return subtle.ConstantTimeCompare(want, got) == 1
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), err
}

// Recovery codes have 128 bits of entropy but avoid mixed-case URL characters,
// making them easier to transcribe from a printed or saved copy.
func newRecoveryCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	value := strings.ToUpper(hex.EncodeToString(b))
	return value[:8] + "-" + value[8:16] + "-" + value[16:24] + "-" + value[24:], nil
}

// Only normalize the new hexadecimal format. Old, case-sensitive base64url
// codes remain valid verbatim, including their hyphens and underscores.
func recoveryCodeHash(code string) string {
	code = strings.TrimSpace(code)
	normalized := strings.Map(func(r rune) rune {
		switch r {
		case '-', ' ', '\t', '\r', '\n':
			return -1
		}
		return r
	}, code)
	if len(normalized) == 32 {
		if _, err := hex.DecodeString(normalized); err == nil {
			code = strings.ToUpper(normalized)
		}
	}
	return digest(code)
}
func digest(value string) string { s := sha256.Sum256([]byte(value)); return hex.EncodeToString(s[:]) }
