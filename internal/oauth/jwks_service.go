package oauth

type JWKSService struct {
	Store *JWKSStore
}

func NewJWKSService(store *JWKSStore) JWKSService {
	return JWKSService{
		Store: store,
	}
}
