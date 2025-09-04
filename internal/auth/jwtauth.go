package auth

import (
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

type Auth struct {
	config        *jwtConfig
	baseTokenAuth *jwtauth.JWTAuth
}

type jwtConfig struct {
	alg string
	key []byte
}

func NewAuth() *Auth {
	newConfig := jwtConfig{alg: "HS256", key: []byte("secret")}
	baseToken := jwtauth.New(newConfig.alg, newConfig.key, nil)
	return &Auth{config: &newConfig, baseTokenAuth: baseToken}
}

func (auth *Auth) CreateUserToken(userId int) (string, error) {
	_, tokenString, err := auth.baseTokenAuth.Encode(map[string]interface{}{"user_id": userId, "exp": time.Now().Add(2 * time.Hour)})
	if err != nil {
		return "", fmt.Errorf("error create user token. %w", err)
	}
	return tokenString, nil
}

func (auth *Auth) GetBaseToken() *jwtauth.JWTAuth {
	return auth.baseTokenAuth
}
