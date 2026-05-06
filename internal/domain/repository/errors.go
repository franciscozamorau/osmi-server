package repository

import "errors"

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrPaymentNotFound = errors.New("payment not found")
)
