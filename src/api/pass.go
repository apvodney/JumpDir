package api

import (
	"github.com/apvodney/JumpDir/api/saltPepper"

	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"strconv"
	"strings"
)

type algoEntry struct {
	hash    func(pass string, args interface{}) (error, string)
	compare func(pass, hash string, args interface{}) (error, bool)
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

func scryptCore(pass string, sp *saltPepper.SaltPepper, args interface{}) (error, string) {
	a, ok := args.(scryptArgs)
	if !ok {
		return errors.New("Can't cast arguments"), ""
	}
	hash, err := scrypt.Key([]byte(pass), sp.SaltPepper, a.N, a.r, a.p, a.keyLen)
	if err != nil {
		return fmt.Errorf("Hash error: %w", err), ""
	}
	return nil, fmt.Sprintf("%s$%s$%s",
		base64.URLEncoding.EncodeToString(sp.Salt),
		strconv.Itoa(sp.PepperIndex),
		base64.URLEncoding.EncodeToString(hash))
}

// Only returns error in very unusual circumstances
func scryptHash(pass string, args interface{}) (error, string) {
	err, sp := saltPepper.NewSaltPepper()
	if err != nil {
		return err, ""
	}
	return scryptCore(pass, sp, args)
}

func scryptCompare(pass, hash1 string, args interface{}) (error, bool) {
	var parseErr error = errors.New("Fatal error, unparsable password hash")
	sh := strings.Split(hash1, "$")
	if len(sh) != 3 {
		return parseErr, false
	}
	salt, err := base64.URLEncoding.DecodeString(sh[0])
	if err != nil {
		return parseErr, false
	}
	pepperIndex, err := strconv.Atoi(sh[1])
	if err != nil {
		return parseErr, false
	}
	err, sp := saltPepper.OldSaltPepper(salt, pepperIndex)
	if err != nil {
		return err, false
	}

	err, hash2 := scryptCore(pass, sp, args)
	if err != nil {
		return err, false
	}
	return nil, hash1 == hash2
}

func (a *Api) passwdBatonRecv() struct{} {
	return <-a.hashLimiter
}

func (a *Api) passwdBatonPass(baton struct{}) {
	a.hashLimiter <- baton
}

func (a *Api) passHash(pass string) (error, string, int32) {
	baton := a.passwdBatonRecv()
	defer a.passwdBatonPass(baton)
	hashAlgo := int32(len(passAlgos) - 1)
	entry := passAlgos[hashAlgo]
	err, passHash := entry.hash(pass, entry.args)
	return err, passHash, hashAlgo
}

func (a *Api) passCompare(pass, hash string, hashAlgo int32) (error, bool) {
	if hashAlgo < 0 || int(hashAlgo) >= len(passAlgos) {
		return errors.New("Fatal error, incorrect algo number"), false
	}
	entry := passAlgos[hashAlgo]
	return entry.compare(pass, hash, entry.args)
}
