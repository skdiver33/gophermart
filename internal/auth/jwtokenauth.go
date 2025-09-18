package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

func (auth *Auth) GetBaseToken() *jwtauth.JWTAuth {
	return auth.baseTokenAuth
}

func (auth *Auth) CreateUserToken(userID int) (string, error) {
	_, tokenString, err := auth.baseTokenAuth.Encode(map[string]interface{}{"user_id": strconv.Itoa(userID), "exp": time.Now().Add(2 * time.Hour)})
	if err != nil {
		return "", fmt.Errorf("error create user token. %w", err)
	}
	return tokenString, nil
}

func (auth *Auth) GetUserIDFromClaims(ctx context.Context) (int, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userIDValue, ok := claims["user_id"]
	if !ok {
		return -1, errors.New("userID not found in JWT claims")
	}
	userIDStr, ok := userIDValue.(string)
	if !ok {
		return -1, errors.New("userID in JWT claims is not string")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return -1, errors.New("error convert user id from string to int")
	}
	return userID, nil
}
