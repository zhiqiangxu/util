package claim

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Verifier for claims
type Verifier struct {
	method    jwt.SigningMethod
	verifyKey *rsa.PublicKey
}

// NewVerifier is ctor for Verifier
func NewVerifier(verifyKey *rsa.PublicKey) (v *Verifier, err error) {
	signingAlgorithm := "RS256"
	method := jwt.GetSigningMethod(signingAlgorithm)
	if method == nil {
		err = fmt.Errorf("invalid signingAlgorithm:%s", method)
		return
	}

	v = &Verifier{method: method, verifyKey: verifyKey}
	return
}

// Verify claims
func (v *Verifier) Verify(tokenString string) (ok bool, values map[string]interface{}) {
	ok, values = verify(tokenString, v.method, v.verifyKey)
	return
}

// GetInt64 for retrieve claim value as int64
func GetInt64(value interface{}) (int64Value int64) {
	switch exp := value.(type) {
	case float64:
		int64Value = int64(exp)
	case json.Number:
		int64Value, _ = exp.Int64()
	}
	return
}

func verify(tokenString string, method jwt.SigningMethod, verifyKey interface{}) (ok bool, values map[string]interface{}) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (key interface{}, err error) {
		if method != token.Method {
			err = fmt.Errorf("invalid signingAlgorithm:%s", token.Method.Alg())
			return
		}
		key = verifyKey
		return
	})
	if err != nil || !token.Valid {
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	expireAt, exists := claims[ExpireATKey]
	if !exists {
		ok = false
		return
	}

	ok = verifyExp(GetInt64(expireAt))
	if !ok {
		return
	}

	values = (map[string]interface{})(claims)
	delete(values, ExpireATKey)
	delete(values, CreatedKey)
	return
}

func verifyExp(exp int64) (ok bool) {
	nowSecond := time.Now().Unix()
	ok = exp > nowSecond
	return
}
