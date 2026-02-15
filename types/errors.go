package types

import "fmt"

var (
	ErrSecretNotFound      = fmt.Errorf("secret not found")
	ErrSecretKeyNotFound   = fmt.Errorf("secret key not found")
	ErrCouldNotFetchSecret = fmt.Errorf("could not fetch secret")
)
