package withdraw

import (
	"regexp"
	"time"
)

type Withdraw struct {
	OrderNumber string    `json:"order,omitempty"`
	UserId      int       `json:"-"`
	Sum         int       `json:"sum,omitempty"`
	UploadData  time.Time `json:"processed_at,omitempty"`
}

func (withdraw *Withdraw) CheckNumber() bool {
	pattern := regexp.MustCompile("^[0-9]+$")
	return pattern.MatchString(withdraw.OrderNumber)
}
