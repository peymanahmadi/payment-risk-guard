package domain

import "errors"

var (
	ErrInvalidAccountID = errors.New("domain: account id must not be empty")
	ErrInvalidAmount = errors.New("domain: amount must be greater than zero")
	ErrNotFound = errors.New("domain: resource not found")
)