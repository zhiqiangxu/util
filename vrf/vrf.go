package vrf

import (
	"crypto"
)

// Signer for vrf
type Signer interface {
	Hash(alpha []byte) (beta [32]byte, proof []byte, err error)
	Public() crypto.PublicKey
}

// Verifier for vrf
type Verifier interface {
	Verify(m, proof []byte, beta [32]byte) (valid bool, err error)
}
