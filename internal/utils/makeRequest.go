package utils

import (
	"io"
	"log"
	"net/http"
	"newsgetter/internal/middleware"
	"time"
)

func MakeRequest(method, path string) (*string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

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

	return &bodyString, nil
}
