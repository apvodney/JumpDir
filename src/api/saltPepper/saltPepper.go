package saltPepper

func constructSaltPepper(salt, pepper []byte) []byte {
	saltPepper := make([]byte, len(salt)+len(pepper))
	saltPepper = append(saltPepper, salt...)
	saltPepper = append(saltPepper, pepper...)
	return saltPepper
}

type SaltPepper struct {
	SaltPepper  []byte
	Salt        []byte
	PepperIndex int
}

func NewSaltPepper() *SaltPepper {
	salt := newSalt()
	pepper, pepperIndex := newPepper()
	return &SaltPepper{
		SaltPepper:  constructSaltPepper(salt, pepper),
		Salt:        salt,
		PepperIndex: pepperIndex,
	}
}

func OldSaltPepper(salt []byte, pepperIndex int) *SaltPepper {
	pepper := oldPepper(pepperIndex)
	return &SaltPepper{
		SaltPepper:  constructSaltPepper(salt, pepper),
		Salt:        salt,
		PepperIndex: pepperIndex,
	}
}
