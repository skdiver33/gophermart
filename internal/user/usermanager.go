package user

import (
	"fmt"
)

type UserManager struct {
	storage       UserStorageInterface
	authenticator AuthInterface
}

func NewUserManager(store UserStorageInterface, auth AuthInterface) *UserManager {
	return &UserManager{storage: store, authenticator: auth}
}

func (manager *UserManager) UserRegister(data *User) (*User, error) {
	//find user in storage if exist return error if not found add user to storage
	user, err := manager.storage.GetUser(data.Login, data.Password)
	if err != nil {
		return nil, fmt.Errorf("error get user from storage %w", err)
	}
	if user != nil {
		return nil, fmt.Errorf("error. User already exist")
	}
	newUser := &User{Login: data.Login, Password: data.Password}

	id, err := manager.storage.AddUser(newUser)
	if err != nil {
		return nil, fmt.Errorf("error add user to storage %w", err)
	}
	newUser.Id = id
	return newUser, nil
}

func (manager *UserManager) UserAuth(data *User) (string, error) {
	//find user in storage if not found - return error else generate and return token
	user, err := manager.storage.GetUser(data.Login, data.Password)
	if err != nil {
		return "", fmt.Errorf("error get user from storage. %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("error. User  with login %v is not register in system", data.Login)
	}
	userAuthToken, err := manager.authenticator.CreateUserToken(user.Id)
	if err != nil {
		return "", fmt.Errorf("error. Cannot create user authentication token. %w", err)
	}
	return userAuthToken, nil
}
