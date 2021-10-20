package postgres

import (
	"crypto/rand"
	"fmt"
)

func tokenGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
