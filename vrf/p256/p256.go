// Package p256 implements a verifiable random function using curve p256.
package p256

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"math/big"

	"github.com/zhiqiangxu/util/vrf"
)

// PublicKey for vrf
type PublicKey struct {
	*ecdsa.PublicKey
}

// PrivateKey for vrf
type PrivateKey struct {
	*ecdsa.PrivateKey
}

var (
	curve  = elliptic.P256()
	params = curve.Params()
)

// H1 hashes m to a curve point
func H1(m []byte) (x, y *big.Int, err error) {
	h := sha512.New()
	var i uint32
	byteLen := (params.BitSize + 7) >> 3
	for x == nil && i < 100 {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		r := []byte{2} // Set point encoding to "compressed", y=0.
		r = h.Sum(r)
		x, y, err = Unmarshal(curve, r[:byteLen+1])
		i++
	}
	return
}

var one = big.NewInt(1)

// H2 hashes m to an integer [1,N-1]
func H2(m []byte) *big.Int {
	// NIST SP 800-90A § A.5.1: Simple discard method.
	byteLen := (params.BitSize + 7) >> 3
	h := sha512.New()
	for i := uint32(0); ; i++ {
		// TODO: Use a NIST specified DRBG.
		h.Reset()
		binary.Write(h, binary.BigEndian, i)
		h.Write(m)
		b := h.Sum(nil)
		k := new(big.Int).SetBytes(b[:byteLen])
		if k.Cmp(new(big.Int).Sub(params.N, one)) == -1 {
			return k.Add(k, one)
		}
	}
}

func (k PrivateKey) Hash(alpha []byte) (beta [32]byte, proof []byte, err error) {
	r, _, _, err := elliptic.GenerateKey(curve, rand.Reader)
	if err != nil {
		return
	}

	ri := new(big.Int).SetBytes(r)

	// H = H1(m)
	Hx, Hy, err := H1(alpha)
	if err != nil {
		return
	}

	// VRF_k(m) = [k]H
	sHx, sHy := curve.ScalarMult(Hx, Hy, k.D.Bytes())
	vrf := elliptic.Marshal(curve, sHx, sHy) // 65 bytes.

	// G is the base point
	// s = H2(G, H, [k]G, VRF, [r]G, [r]H)
	rGx, rGy := curve.ScalarBaseMult(r)
	rHx, rHy := curve.ScalarMult(Hx, Hy, r)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, k.PublicKey.X, k.PublicKey.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, rGx, rGy))
	b.Write(elliptic.Marshal(curve, rHx, rHy))
	s := H2(b.Bytes())

	// t = r−s*k mod N
	t := new(big.Int).Sub(ri, new(big.Int).Mul(s, k.D))
	t.Mod(t, params.N)

	// beta = H(vrf)
	beta = sha256.Sum256(vrf)

	// Write s, t, and vrf to a proof blob. Also write leading zeros before s and t
	// if needed.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(s.Bytes())))
	buf.Write(s.Bytes())
	buf.Write(make([]byte, 32-len(t.Bytes())))
	buf.Write(t.Bytes())
	buf.Write(vrf)
	proof = buf.Bytes()

	return
}

// Public returns the corresponding public key as bytes.
func (k PrivateKey) Public() crypto.PublicKey {
	return &k.PublicKey
}

func (pk PublicKey) Verify(m, proof []byte, beta [32]byte) (valid bool, err error) {
	// verifier checks that s == H2(m, [t]G + [s]([k]G), [t]H1(m) + [s]VRF_k(m))
	if got, want := len(proof), 64+65; got != want {
		return
	}

	// Parse proof into s, t, and vrf.
	s := proof[0:32]
	t := proof[32:64]
	vrf := proof[64 : 64+65]

	uHx, uHy := elliptic.Unmarshal(curve, vrf)
	if uHx == nil {
		return
	}

	// [t]G + [s]([k]G) = [t+ks]G
	tGx, tGy := curve.ScalarBaseMult(t)
	ksGx, ksGy := curve.ScalarMult(pk.X, pk.Y, s)
	tksGx, tksGy := curve.Add(tGx, tGy, ksGx, ksGy)

	// H = H1(m)
	// [t]H + [s]VRF = [t+ks]H
	Hx, Hy, err := H1(m)
	if err != nil {
		return
	}
	tHx, tHy := curve.ScalarMult(Hx, Hy, t)
	sHx, sHy := curve.ScalarMult(uHx, uHy, s)
	tksHx, tksHy := curve.Add(tHx, tHy, sHx, sHy)

	//   H2(G, H, [k]G, VRF, [t]G + [s]([k]G), [t]H + [s]VRF)
	// = H2(G, H, [k]G, VRF, [t+ks]G, [t+ks]H)
	// = H2(G, H, [k]G, VRF, [r]G, [r]H)
	var b bytes.Buffer
	b.Write(elliptic.Marshal(curve, params.Gx, params.Gy))
	b.Write(elliptic.Marshal(curve, Hx, Hy))
	b.Write(elliptic.Marshal(curve, pk.X, pk.Y))
	b.Write(vrf)
	b.Write(elliptic.Marshal(curve, tksGx, tksGy))
	b.Write(elliptic.Marshal(curve, tksHx, tksHy))
	h2 := H2(b.Bytes())

	// Left pad h2 with zeros if needed. This will ensure that h2 is padded
	// the same way s is.
	var buf bytes.Buffer
	buf.Write(make([]byte, 32-len(h2.Bytes())))
	buf.Write(h2.Bytes())

	if !hmac.Equal(s, buf.Bytes()) {
		return
	}

	valid = true
	return
}

// GeneratePair generates a fresh signer/verifier pair for this VRF
func GeneratePair() (k vrf.Signer, pk vrf.Verifier, err error) {
	key, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return
	}

	k, pk = PrivateKey{PrivateKey: key}, PublicKey{PublicKey: &key.PublicKey}
	return
}

var (
	errPointNotOnCurve = errors.New("point not on curve")
	errNoPEMFound      = errors.New("no pem found")
	errWrongKeyType    = errors.New("wrong key type")
)

// NewVRFSigner creates a signer object from a private key.
func NewVRFSigner(key *ecdsa.PrivateKey) (k vrf.Signer, err error) {
	if *(key.Params()) != *curve.Params() {
		err = errPointNotOnCurve
		return
	}
	if !curve.IsOnCurve(key.X, key.Y) {
		err = errPointNotOnCurve
		return
	}
	k = PrivateKey{PrivateKey: key}
	return
}

// NewVRFVerifier creates a verifier object from a public key.
func NewVRFVerifier(pubkey *ecdsa.PublicKey) (pk vrf.Verifier, err error) {
	if *(pubkey.Params()) != *curve.Params() {
		err = errPointNotOnCurve
		return
	}
	if !curve.IsOnCurve(pubkey.X, pubkey.Y) {
		err = errPointNotOnCurve
		return
	}
	pk = PublicKey{PublicKey: pubkey}
	return
}

// NewVRFSignerFromRawKey returns the private key from a raw private key bytes.
func NewVRFSignerFromRawKey(b []byte) (k vrf.Signer, err error) {
	key, err := x509.ParseECPrivateKey(b)
	if err != nil {
		return
	}
	k, err = NewVRFSigner(key)
	return
}

// NewVRFSignerFromPEM creates a vrf private key from a PEM data structure.
func NewVRFSignerFromPEM(b []byte) (k vrf.Signer, err error) {
	p, _ := pem.Decode(b)
	if p == nil {
		err = errNoPEMFound
		return
	}
	k, err = NewVRFSignerFromRawKey(p.Bytes)
	return
}

// NewVRFVerifierFromRawKey returns the public key from a raw public key bytes.
func NewVRFVerifierFromRawKey(b []byte) (pk vrf.Verifier, err error) {
	key, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return
	}
	pubkey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		err = errWrongKeyType
		return
	}
	pk, err = NewVRFVerifier(pubkey)
	return
}
