package store

import (
	"strconv"

	"github.com/google/uuid"
)

func newUUID() string {
	return uuid.New().String()
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
