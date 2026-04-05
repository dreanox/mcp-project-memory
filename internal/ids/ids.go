package ids

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// New returns a short unique id (time + random).
func New() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return time.Now().UTC().Format("20060102") + "-" + hex.EncodeToString(b[:])
}
