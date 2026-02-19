package globalerrors

import "errors"

var (
	ErrorNotFound = errors.New("subscription not found")
)
