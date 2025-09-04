package storage

import (
	"errors"

	user "github.com/skdiver33/gophermart/internal/user"
)

var count int

type UserMemStorage struct {
	users map[int]user.User
}

func NewUserMemStorage() *UserMemStorage {
	newStorage := make(map[int]user.User)
	return &UserMemStorage{users: newStorage}
}

func init() {
	count = 0
}

func (storage *UserMemStorage) AddUser(user *user.User) (int, error) {
	user.Id = count
	storage.users[count] = *user
	count++
	return user.Id, nil

}

func (storage *UserMemStorage) GetUser(login string, password string) (*user.User, error) {
	for _, val := range storage.users {
		if val.Login == login && val.Password == password {
			return &val, nil
		}
	}
	return nil, errors.New("user not found")
}
