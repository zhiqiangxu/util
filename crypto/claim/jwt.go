package claim

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// JWT is both signer and verifier
type JWT struct {
	expire time.Duration
	method jwt.SigningMethod
	secret []byte
}

// NewJWT is ctor for JWT
func NewJWT(expire time.Duration, secret []byte) (j *JWT, err error) {
	signingAlgorithm := "HS256"
	method := jwt.GetSigningMethod(signingAlgorithm)
	if method == nil {
		err = fmt.Errorf("invalid signingAlgorithm:%s", method)
		return
	}
	j = &JWT{expire: expire, method: method, secret: secret}
	return
}

// Sign claims
func (j *JWT) Sign(values map[string]interface{}) (tokenString string, err error) {

	tokenString, err = sign(values, j.expire, j.method, j.secret)

	return
}

// Verify claims
func (j *JWT) Verify(tokenString string) (ok bool, values map[string]interface{}) {
	ok, values = verify(tokenString, j.method, j.secret)
	return
}
