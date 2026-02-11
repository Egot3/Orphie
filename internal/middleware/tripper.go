package middleware

import (
	"net/http"
	"newsgetter/internal/types"
)

func UseTripper(client *http.Client, before func(*http.Request) error, after func(*http.Request, *http.Response, error) error) {
	if client.Transport == nil {
		client.Transport = http.DefaultTransport
	}

	client.Transport = &types.MiddlewareTripper{
		Next:   client.Transport,
		Before: before,
		After:  after,
	}
}
