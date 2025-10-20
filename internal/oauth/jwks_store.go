package oauth

import (
	"context"
	"errors"

	"github.com/freekieb7/smauth/internal/database"
)

type JWKSStore struct {
	DB *database.Database
}

func NewJWKSStore(db *database.Database) JWKSStore {
	return JWKSStore{DB: db}
}

func (s JWKSStore) GetJWKS(ctx context.Context) (JWKS, error) {
	var jwks JWKS
	rows, err := s.DB.Conn.Query(ctx, "SELECT id, kid, kty, alg, use, n, e, x5c, x5t FROM tbl_jwks")
	if err != nil {
		return JWKS{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var j JWK
		if err := rows.Scan(&j.ID, &j.KID, &j.KTY, &j.ALG, &j.USE, &j.N, &j.E, &j.X5C, &j.X5T); err != nil {
			return JWKS{}, err
		}
		jwks.Keys = append(jwks.Keys, j)
	}
	return jwks, nil
}

func (s JWKSStore) SaveJWK(ctx context.Context, jwk JWK) error {
	if jwk.IsNew {
		_, err := s.DB.Conn.Exec(ctx, `
			INSERT INTO tbl_jwks (id, origin, kid, kty, alg, use, n, e, x5c, x5t)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, jwk.ID, jwk.Origin, jwk.KID, jwk.KTY, jwk.ALG, jwk.USE, jwk.N, jwk.E, jwk.X5C, jwk.X5T)
		return err
	}

	return errors.New("updating a jwk is not supported")
}
