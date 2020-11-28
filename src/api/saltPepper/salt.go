package saltPepper

import (
	"github.com/apvodney/JumpDir/api/fatalError"

	"crypto/rand"
	"fmt"
)

func newSalt() []byte {
	salt := make([]byte, 16)

	_, err := rand.Read(salt)
	if err != nil {
		fatalError.Panic(fmt.Errorf("Error getting salt: %w", err))
	}

	return salt
}
