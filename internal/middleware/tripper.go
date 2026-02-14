package middleware

import (
	"net/http"
	"newsgetter/internal/types"
)

func UseTripper(client *types.Client, before func(*http.Request) error, after func(*http.Request, *http.Response, error) error) {
	if client.Base.Transport == nil {
		client.Base.Transport = http.DefaultTransport
	}

	client.Base.Transport = &types.MiddlewareTripper{
		Next:   client.Base.Transport,
		Before: before,
		After:  after,
	}
}
