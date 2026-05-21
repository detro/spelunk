package types

import "fmt"

var (
	ErrCouldNotFetchSecret = fmt.Errorf("could not fetch secret")
	ErrInvalidLocation     = fmt.Errorf("invalid secret location format")
	ErrSecretKeyNotFound   = fmt.Errorf("secret key not found")
	ErrSecretNotFound      = fmt.Errorf("secret not found")
)
