package request

import (
	"io"
	"log"
	"net/http"
	"newsgetter/internal/client"
	"newsgetter/internal/middleware"
	"newsgetter/internal/types"
	"time"
)

func MakeRequest(method, path string) (*types.Response, error) {
	client := client.NewClient(&http.Client{Timeout: 5 * time.Second})

	before, after := middleware.TraceTripperMiddleware()
	middleware.UseTripper(client, before, after)

	req, _ := http.NewRequest(method, path, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Body parsing error: %v", err)
	}

	bodyString := string(bodyBytes)

	return &types.Response{
		Body:       bodyString,
		Method:     resp.Request.Method,
		Path:       resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
	}, nil
}
