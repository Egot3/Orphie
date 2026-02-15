package types

import (
	"crypto/sha256"
)

type Response struct {
	Path       string
	Method     string
	Headers    map[string][]string
	StatusCode int
	Body       string
}

func (r Response) Hash() [32]uint8 {

	hash := sha256.Sum256([]byte(r.Body + r.Method))

	return hash
}
