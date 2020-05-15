package signature

import (
	"crypto/sha512"
	"hash"
)

func getHasher() hash.Hash {
	return sha512.New()
}
