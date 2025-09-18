package order

import (
	"regexp"
	"time"
)

const (
	OrderStatusNew        = "NEW"
	OrderStautsProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Order struct {
	Number     string    `json:"number,omitempty"`
	UserId     int       `json:"-"`
	Status     string    `json:"status,omitempty"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadData time.Time `json:"uploaded_at,omitempty"`
}

func (order *Order) CheckNumber() bool {
	pattern := regexp.MustCompile("^[0-9]+$")
	return pattern.MatchString(order.Number)
}
