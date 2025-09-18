package user

import "context"

type UserManagerInterface interface {
	UserRegister(ctx context.Context, data *User) (*User, error) //for successful add user to storage, setup user id and return it
	UserAuth(ctx context.Context, data *User) (string, error)    // for succesfull return JWT
}

type UserStorageInterface interface {
	AddUser(ctx context.Context, user *User) (int, error)
	GetUser(ctx context.Context, login string, password string) (*User, error)
}

type AuthInterface interface {
	CreateUserToken(userID int) (string, error)
	GetUserIDFromClaims(ctx context.Context) (int, error)
}
