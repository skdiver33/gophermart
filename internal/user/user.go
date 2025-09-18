package user

type User struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	ID       int    `json:"id,omitempty"`
}
