package user

type UserManagerInterface interface {
	UserRegister(data *User) (*User, error) //for successful add user to storage, setup user id and return it
	UserAuth(data *User) (string, error)    // for succesfull return JWT
}

type UserStorageInterface interface {
	AddUser(user *User) (int, error)
	GetUser(login string, password string) (*User, error)
}

type AuthInterface interface {
	CreateUserToken(userId int) (string, error)
}
