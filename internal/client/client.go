package client

import (
	"net/http"
	"newsgetter/internal/types"
	"time"
)

func NewClient(base *http.Client) *types.Client {
	if base == nil {
		base = &http.Client{Timeout: 10 * time.Second}
	}
	return &types.Client{Base: base}
}
