package claim

import (
	"io/ioutil"
	"testing"
	"time"

	"reflect"

	"gopkg.in/dgrijalva/jwt-go.v3"
	"gotest.tools/assert"
)

func Test(t *testing.T) {

	// test signer and verifier
	{
		// openssl genrsa -out test.rsa 2048
		signBytes, err := ioutil.ReadFile("data/test.rsa")
		assert.Assert(t, err == nil)
		// openssl rsa -in test.rsa -pubout > test.rsa.pub
		verifyBytes, err := ioutil.ReadFile("data/test.rsa.pub")
		assert.Assert(t, err == nil)

		signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
		assert.Assert(t, err == nil)
		verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
		assert.Assert(t, err == nil)

		s, err := NewSigner(time.Hour, signKey)
		assert.Assert(t, err == nil)

		v, err := NewVerifier(verifyKey)
		assert.Assert(t, err == nil)

		values := map[string]interface{}{
			"c1": "v1",
			"c2": "v2",
		}

		tokenString, err := s.Sign(values)
		assert.Assert(t, err == nil, err)

		ok, valuesVerified := v.Verify(tokenString)
		assert.Assert(t, ok && reflect.DeepEqual(valuesVerified, values))
	}

	// test jwt
	{
		j, err := NewJWT(time.Hour*24, []byte("sec"))
		assert.Assert(t, err == nil)

		values := map[string]interface{}{
			"c1": "v1",
			"c2": "v2",
		}
		tokenString, err := j.Sign(values)
		assert.Assert(t, err == nil)

		ok, valuesVerified := j.Verify(tokenString)
		assert.Assert(t, ok && reflect.DeepEqual(valuesVerified, values))
	}

}
