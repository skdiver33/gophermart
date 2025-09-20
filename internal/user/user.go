package user

import (
	"crypto/sha256"
	"encoding/base64"
)

type User struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	ID       int    `json:"id,omitempty"`
}

func (user *User) CryptPasswd() {
	h := sha256.New()
	h.Write([]byte(user.Password))
	dst := h.Sum(nil)
	user.Password = base64.RawStdEncoding.EncodeToString(dst)
}
