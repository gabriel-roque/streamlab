package video

import (
	"crypto/rand"
	"encoding/hex"
)

func NewID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return prefix + "-00000000"
	}
	return prefix + "-" + hex.EncodeToString(b)
}
