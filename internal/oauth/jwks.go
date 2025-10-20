package oauth

import (
	"crypto/rsa"
	"encoding/base64"

	"github.com/google/uuid"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	ID     uuid.UUID `json:"-"`
	Origin string    `json:"-"` // e.g., 'microsoft', 'oauth_provider_x'
	KID    string    `json:"kid"`
	KTY    string    `json:"kty"`
	ALG    string    `json:"alg"`
	USE    string    `json:"use"`
	N      string    `json:"n"`
	E      string    `json:"e"`
	X5C    []string  `json:"x5c"`
	X5T    string    `json:"x5t"`
	IsNew  bool      `json:"-"`
}

func (j JWK) RSAPublicKey() (*rsa.PublicKey, error) {
	if j.KTY != "RSA" {
		return nil, nil
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, err
	}

	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	pubKey := &rsa.PublicKey{
		N: new(rsa.PublicKey).N.SetBytes(nBytes),
		E: eInt,
	}

	return pubKey, nil
}
