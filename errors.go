package client

import (
	"errors"
)

var (
	// ErrUnsupportedContentEncoding is an error thrown when the response header Content-Encoding is not supported
	ErrUnsupportedContentEncoding = errors.New("Unsupported Content-Encoding value")
)
