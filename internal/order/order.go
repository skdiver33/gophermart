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
	UserID     int       `json:"-"`
	Status     string    `json:"status,omitempty"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadData time.Time `json:"uploaded_at,omitempty"`
}

func (order *Order) CheckNumber() bool {
	pattern := regexp.MustCompile("^[0-9]+$")
	return pattern.MatchString(order.Number)
}

func (order *Order) LunaCheck() bool {

	pattern := regexp.MustCompile("^[0-9]+$")
	if !pattern.MatchString(order.Number) {
		return false
	}

	sum := 0
	nDigits := len(order.Number)
	parity := nDigits % 2
	for i := 0; i < nDigits; i++ {
		digit := int(order.Number[i]) - '0'
		if digit%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	ret := false
	if sum%10 == 0 {
		ret = true
	}
	return ret
}
