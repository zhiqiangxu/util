package p256

import (
	"crypto/rand"
	"testing"
)

func TestH1(t *testing.T) {
	for i := 0; i < 10000; i++ {
		m := make([]byte, 100)
		if _, err := rand.Read(m); err != nil {
			t.Fatalf("Failed generating random message: %v", err)
		}
		x, y, err := H1(m)
		if err != nil {
			t.Errorf("H1(%v)=%v, want curve point", m, x)
		}
		if got := curve.Params().IsOnCurve(x, y); !got {
			t.Errorf("H1(%v)=[%v, %v], is not on curve", m, x, y)
		}
	}
}

func TestH2(t *testing.T) {
	l := 32
	for i := 0; i < 10000; i++ {
		m := make([]byte, 100)
		if _, err := rand.Read(m); err != nil {
			t.Fatalf("Failed generating random message: %v", err)
		}
		x := H2(m)
		if got := len(x.Bytes()); got < 1 || got > l {
			t.Errorf("len(h2(%v)) = %v, want: 1 <= %v <= %v", m, got, got, l)
		}
	}
}

func TestVRF(t *testing.T) {
	k, pk, err := GeneratePair()
	if err != nil {
		t.Fatalf("Failed generating key pairs: %v", err)
	}

	m1 := []byte("data1")
	m2 := []byte("data2")
	m3 := []byte("data2")
	beta1, proof1, err := k.Hash(m1)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	beta2, proof2, err := k.Hash(m2)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	beta3, proof3, err := k.Hash(m3)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}
	for _, tc := range []struct {
		m     []byte
		beta  [32]byte
		proof []byte
		valid bool
	}{
		{m1, beta1, proof1, true},
		{m2, beta2, proof2, true},
		{m3, beta3, proof3, true},
		{m3, beta3, proof2, true},
		{m3, beta3, proof1, false},
	} {
		valid, err := pk.Verify(tc.m, tc.proof, tc.beta)
		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		}
		if got, want := valid, tc.valid; got != want {
			t.Errorf("Verify(%s, %x): %v, want %v", tc.m, tc.proof, got, want)
		}
	}
}
