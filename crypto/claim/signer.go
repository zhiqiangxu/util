package claim

import (
	"crypto/rsa"
	"fmt"
	"time"

	"gopkg.in/dgrijalva/jwt-go.v3"
)

// Signer for claimer
type Signer struct {
	expire  time.Duration
	method  jwt.SigningMethod
	signKey *rsa.PrivateKey
}

// NewSigner is ctor for Signer
func NewSigner(expire time.Duration, signKey *rsa.PrivateKey) (s *Signer, err error) {
	signingAlgorithm := "RS256"
	method := jwt.GetSigningMethod(signingAlgorithm)
	if method == nil {
		err = fmt.Errorf("invalid signingAlgorithm:%s", method)
		return
	}
	s = &Signer{expire: expire, method: method, signKey: signKey}
	return
}

const (
	// ExpireATKey for expire_at
	ExpireATKey = "expire_at"
	// CreatedKey for created
	CreatedKey = "created"
)

// Sign claims
func (s *Signer) Sign(values map[string]interface{}) (tokenString string, err error) {

	tokenString, err = sign(values, s.expire, s.method, s.signKey)
	return
}

func sign(values map[string]interface{}, expire time.Duration, method jwt.SigningMethod, signKey interface{}) (tokenString string, err error) {
	claims := jwt.MapClaims{
		ExpireATKey: time.Now().Add(expire).Unix(),
		CreatedKey:  time.Now().Unix(),
	}
	for k, v := range values {
		if _, ok := claims[k]; ok {
			err = fmt.Errorf("%s is reserved for claims", k)
			return
		}
		claims[k] = v
	}
	token := jwt.NewWithClaims(method, claims)
	// Sign and get the complete encoded token as a string using the secret
	tokenString, err = token.SignedString(signKey)
	return
}
