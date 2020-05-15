package signature

import (
	"testing"

	"gotest.tools/assert"
)

func TestSignature(t *testing.T) {
	pri, pub, err := GenerateKeypair()

	assert.Assert(t, err == nil)

	msg := []byte("hello msg")
	sig, err := Sign(pri, msg)
	assert.Assert(t, err == nil)

	ok, err := Verify(pub, msg, sig)
	assert.Assert(t, err == nil && ok)

	sigbytes, err := sig.Marshal()
	assert.Assert(t, err == nil)
	var sig2 Signature
	err = sig2.Unmarshal(sigbytes)
	assert.Assert(t, err == nil)
	ok, err = Verify(pub, msg, &sig2)
	assert.Assert(t, err == nil && ok)
}
