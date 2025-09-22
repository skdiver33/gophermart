package user

import (
	"crypto/sha256"
	"encoding/base64"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	ID       int    `json:"-"`
}

func (user *User) CryptPasswd() {
	h := sha256.New()
	h.Write([]byte(user.Password))
	dst := h.Sum(nil)
	user.Password = base64.RawStdEncoding.EncodeToString(dst)
}
