package jks

import "errors"

var (
	ErrNoPassword   = errors.New("password is not set for store")
	ErrInvalidAlias = errors.New("alias must not be an empty string")
)
