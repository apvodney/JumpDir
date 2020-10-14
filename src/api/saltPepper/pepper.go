package saltPepper

import "errors"

func newPepper() ([]byte, int) {
	newest := len(peppers) - 1
	return peppers[newest], newest
}

func oldPepper(i int) (error, []byte) {
	if i < 0 || i >= len(peppers) {
		return errors.New("Invalid index"), nil
	}
	return nil, peppers[i]
}
