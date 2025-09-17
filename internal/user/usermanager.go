package user

import (
	"context"
	"errors"
	"fmt"
)

type UserManager struct {
	Storage       UserStorageInterface
	Authenticator AuthInterface
}

func NewUserManager(store UserStorageInterface, auth AuthInterface) *UserManager {
	return &UserManager{Storage: store, Authenticator: auth}
}

var (
	ErrInternal             = errors.New("internal error")
	ErrUserAlreadyExist     = errors.New("error. User already exist")
	ErrUserWithCredNotFound = errors.New("error. User with login and password not found")
)

func (manager *UserManager) UserRegister(ctx context.Context, data *User) (*User, error) {
	//find user in storage if exist return error if not found add user to storage
	user, err := manager.Storage.GetUser(ctx, data.Login, data.Password)
	if err != nil {
		return nil, fmt.Errorf("error get user from storage %w,%w", ErrInternal, err)
	}
	if user != nil {
		return nil, ErrUserAlreadyExist
	}
	newUser := &User{Login: data.Login, Password: data.Password}

	id, err := manager.Storage.AddUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("error add user to storage %w,%w", ErrInternal, err)
	}
	newUser.Id = id
	return newUser, nil
}

func (manager *UserManager) UserAuth(ctx context.Context, data *User) (string, error) {
	//find user in storage if not found - return error else generate and return token
	user, err := manager.Storage.GetUser(ctx, data.Login, data.Password)
	if err != nil {
		return "", fmt.Errorf("error get user from storage. %w %w", ErrInternal, err)
	}
	if user == nil {
		return "", ErrUserWithCredNotFound
	}
	userAuthToken, err := manager.Authenticator.CreateUserToken(user.Id)
	if err != nil {
		return "", fmt.Errorf("error. Cannot create user authentication token. %w %w", ErrInternal, err)
	}
	return userAuthToken, nil
}
