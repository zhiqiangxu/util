package vrf

// Signer for vrf
type Signer interface {
	Hash(alpha []byte) (beta [32]byte, proof []byte, err error)
	Public() Verifier
}

// Verifier for vrf
type Verifier interface {
	Verify(alpha, proof []byte, beta [32]byte) (valid bool, err error)
}
