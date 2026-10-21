package localauth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

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
func seal(secret, plaintext string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	b, e := aes.NewCipher(key[:])
	if e != nil {
		return "", e
	}
	a, e := cipher.NewGCM(b)
	if e != nil {
		return "", e
	}
	nonce := make([]byte, a.NonceSize())
	if _, e = rand.Read(nonce); e != nil {
		return "", e
	}
	return base64.RawStdEncoding.EncodeToString(a.Seal(nonce, nonce, []byte(plaintext), []byte("telefonia-mfa-v1"))), nil
}
func unseal(secret, encoded string) (string, error) {
	key := sha256.Sum256([]byte(secret))
	b, e := aes.NewCipher(key[:])
	if e != nil {
		return "", e
	}
	a, e := cipher.NewGCM(b)
	if e != nil {
		return "", e
	}
	data, e := base64.RawStdEncoding.DecodeString(encoded)
	if e != nil || len(data) < a.NonceSize() {
		return "", fmt.Errorf("invalid MFA secret")
	}
	plain, e := a.Open(nil, data[:a.NonceSize()], data[a.NonceSize():], []byte("telefonia-mfa-v1"))
	return string(plain), e
}

func totp(secret string, step int64) string {
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return ""
	}
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(step))
	mac := hmac.New(sha1.New, key)
	mac.Write(counter[:])
	sum := mac.Sum(nil)
	off := sum[len(sum)-1] & 15
	n := binary.BigEndian.Uint32(sum[off:off+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", n%1000000)
}
func verifyTOTP(secret, code string, now time.Time, lastStep int64) (int64, bool) {
	if len(code) != 6 {
		return 0, false
	}
	if _, e := strconv.Atoi(code); e != nil {
		return 0, false
	}
	step := now.Unix() / 30
	for _, offset := range []int64{0, -1, 1} {
		s := step + offset
		if s > lastStep && subtle.ConstantTimeCompare([]byte(totp(secret, s)), []byte(code)) == 1 {
			return s, true
		}
	}
	return 0, false
}
