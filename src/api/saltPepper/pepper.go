package saltPepper

import (
	"github.com/apvodney/JumpDir/api/fatalError"

	"errors"
)

func newPepper() ([]byte, int) {
	newest := len(peppers) - 1
	return peppers[newest], newest
}

func oldPepper(i int) []byte {
	if i < 0 || i >= len(peppers) {
		fatalError.Panic(errors.New("Invalid pepper index"))
	}
	return peppers[i]
}
