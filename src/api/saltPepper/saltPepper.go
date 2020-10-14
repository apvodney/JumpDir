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

func NewSaltPepper() (error, *SaltPepper) {
	err, salt := newSalt()
	if err != nil {
		return err, nil
	}
	pepper, pepperIndex := newPepper()
	return nil, &SaltPepper{
		SaltPepper:  constructSaltPepper(salt, pepper),
		Salt:        salt,
		PepperIndex: pepperIndex,
	}
}

func OldSaltPepper(salt []byte, pepperIndex int) (error, *SaltPepper) {
	err, pepper := oldPepper(pepperIndex)
	if err != nil {
		return err, nil
	}
	return nil, &SaltPepper{
		SaltPepper:  constructSaltPepper(salt, pepper),
		Salt:        salt,
		PepperIndex: pepperIndex,
	}
}
