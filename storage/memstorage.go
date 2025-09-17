package storage

import (
	order "github.com/skdiver33/gophermart/internal/order"
	user "github.com/skdiver33/gophermart/internal/user"
)

var count int

type UserMemStorage struct {
	users  map[int]user.User
	orders map[string]order.Order
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
	return nil, nil
}

func (storage *UserMemStorage) AddOrder(order *order.Order) error {

	return nil
}

func (storage *UserMemStorage) GetOrder(number string) (*order.Order, error) {
	return nil, nil
}
