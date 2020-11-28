package api

import (
	"github.com/apvodney/JumpDir/api/fatalError"
	"github.com/apvodney/JumpDir/api/saltPepper"

	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"strconv"
	"strings"
)

type algoEntry struct {
	hash    func(pass string, args interface{}) string
	compare func(pass, hash string, args interface{}) bool
	args    interface{}
}

var passAlgos []algoEntry = []algoEntry{
	{hash: scryptHash, compare: scryptCompare, args: scryptArgs{N: 32768, r: 8, p: 1, keyLen: 32}},
}

type scryptArgs struct {
	N      int
	r      int
	p      int
	keyLen int
}

func scryptCore(pass string, sp *saltPepper.SaltPepper, args interface{}) string {
	a, ok := args.(scryptArgs)
	if !ok {
		fatalError.Panic(errors.New("Can't cast arguments"))
	}
	hash, err := scrypt.Key([]byte(pass), sp.SaltPepper, a.N, a.r, a.p, a.keyLen)
	if err != nil {
		fatalError.Panic(fmt.Errorf("Hash error: %w", err))
	}
	return fmt.Sprintf("%s$%s$%s",
		base64.URLEncoding.EncodeToString(sp.Salt),
		strconv.Itoa(sp.PepperIndex),
		base64.URLEncoding.EncodeToString(hash))
}

func scryptHash(pass string, args interface{}) string {
	sp := saltPepper.NewSaltPepper()
	return scryptCore(pass, sp, args)
}

func scryptCompare(pass, hash1 string, args interface{}) bool {
	var parseErr error = errors.New("Fatal error, unparsable password hash")
	sh := strings.Split(hash1, "$")
	if len(sh) != 3 {
		fatalError.Panic(parseErr)
	}
	salt, err := base64.URLEncoding.DecodeString(sh[0])
	if err != nil {
		fatalError.Panic(parseErr)
	}
	pepperIndex, err := strconv.Atoi(sh[1])
	if err != nil {
		fatalError.Panic(parseErr)
	}
	sp := saltPepper.OldSaltPepper(salt, pepperIndex)

	hash2 := scryptCore(pass, sp, args)
	return hash1 == hash2
}

func (a *Api) passwdBatonRecv() struct{} {
	return <-a.hashLimiter
}

func (a *Api) passwdBatonPass(baton struct{}) {
	a.hashLimiter <- baton
}

func (a *Api) passHash(pass string) (string, int32) {
	baton := a.passwdBatonRecv()
	defer a.passwdBatonPass(baton)
	hashAlgo := int32(len(passAlgos) - 1)
	entry := passAlgos[hashAlgo]
	passHash := entry.hash(pass, entry.args)
	return passHash, hashAlgo
}

func (a *Api) passCompare(pass, hash string, hashAlgo int32) bool {
	if hashAlgo < 0 || int(hashAlgo) >= len(passAlgos) {
		fatalError.Panic(errors.New("Fatal error, incorrect algo number"))
	}
	entry := passAlgos[hashAlgo]
	return entry.compare(pass, hash, entry.args)
}
