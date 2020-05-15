package signature

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// Signature ...
type Signature struct {
	r *big.Int
	s *big.Int
}

// Marshal ...
func (sig *Signature) Marshal() (bytes []byte, err error) {
	size := (curve.Params().BitSize + 7) >> 3
	bytes = make([]byte, size*2)

	r := sig.r.Bytes()
	s := sig.s.Bytes()
	copy(bytes[size-len(r):], r)
	copy(bytes[size*2-len(s):], s)
	return
}

// Unmarshal ...
func (sig *Signature) Unmarshal(data []byte) (err error) {
	length := len(data)
	if length&1 != 0 {
		err = errors.New("invalid length")
		return
	}

	sig.r = new(big.Int).SetBytes(data[0 : length/2])
	sig.s = new(big.Int).SetBytes(data[length/2:])
	return
}

// Sign ...
func Sign(privateKey crypto.PrivateKey, msg []byte) (sig *Signature, err error) {
	switch key := privateKey.(type) {
	case *ecdsa.PrivateKey:
		h := getHasher()
		_, err = h.Write(msg)
		if err != nil {
			return
		}
		digest := h.Sum(nil)

		var r, s *big.Int
		r, s, err = ecdsa.Sign(rand.Reader, key, digest)
		if err != nil {
			return
		}
		sig = &Signature{r: r, s: s}
	default:
		err = fmt.Errorf("only ecdsa supported")
	}
	return
}

// Verify ...
func Verify(publicKey crypto.PublicKey, msg []byte, sig *Signature) (ok bool, err error) {
	switch key := publicKey.(type) {
	case *ecdsa.PublicKey:
		h := getHasher()
		_, err = h.Write(msg)
		if err != nil {
			return
		}
		digest := h.Sum(nil)

		ok = ecdsa.Verify(key, digest, sig.r, sig.s)
	default:
		err = fmt.Errorf("only ecdsa supported")
	}
	return
}

var (
	curve = elliptic.P256()
)

// GenerateKeypair ...
func GenerateKeypair() (privateKey crypto.PrivateKey, publicKey crypto.PublicKey, err error) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return
	}

	privateKey = key
	publicKey = &key.PublicKey
	return

}
