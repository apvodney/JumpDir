package saltPepper

import (
	"crypto/rand"
	"fmt"
)

func newSalt() (error, []byte) {
	salt := make([]byte, 16)

	_, err := rand.Read(salt)
	if err != nil {
		return fmt.Errorf("Error getting salt: %w", err), nil
	}

	return nil, salt
}
