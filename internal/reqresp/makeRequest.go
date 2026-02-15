package reqresp

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
	client := client.NewClient(&http.Client{Timeout: 10 * time.Second})

	before, after := middleware.TraceTripperMiddleware()
	middleware.UseTripper(client, before, after)

	req, _ := http.NewRequest(method, path, nil)

	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	req.Header.Set("User-Agent", "orphie/0.1.1") //lawful neutral

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

	//log.Printf("%v\n\n%v", bodyString, resp.Header)

	return &types.Response{
		Body:       bodyString,
		Method:     resp.Request.Method,
		Path:       resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}, nil
}
