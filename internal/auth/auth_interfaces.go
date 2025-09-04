package auth

import "github.com/go-chi/jwtauth/v5"

type JWTManager interface {
	GetBaseToken() *jwtauth.JWTAuth
}
