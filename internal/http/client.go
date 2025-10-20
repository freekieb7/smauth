package http

import (
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
}

func NewClient(timeout time.Duration) Client {
	return Client{
		&http.Client{
			Timeout: timeout,
		},
	}
}

func (c Client) Get(url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.Client.Do(req)
}
