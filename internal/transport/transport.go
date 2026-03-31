// Package transport handles HTTP communication with the YouTube Music API.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	YTMDomain  = "https://music.youtube.com"
	YTMBaseAPI = YTMDomain + "/youtubei/v1/"
	YTMParams  = "?alt=json"
	UserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0"
)

// Client wraps an http.Client and holds base headers/cookies.
type Client struct {
	http    *http.Client
	headers map[string]string
}

// NewClient creates a Client with a 30-second timeout.
func NewClient(headers map[string]string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		headers: headers,
	}
}

// Post sends a POST request to endpoint with JSON body and returns decoded JSON.
func (c *Client) Post(ctx context.Context, endpoint string, params string, body map[string]any, extraHeaders map[string]string) (map[string]any, error) {
	url := YTMBaseAPI + endpoint + YTMParams + params

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("transport: marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("transport: create request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	req.AddCookie(&http.Cookie{Name: "SOCS", Value: "CAI"})

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transport: execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transport: read response body: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("transport: decode response JSON: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("server returned HTTP %d: %s", resp.StatusCode, resp.Status)
		if errObj, ok := result["error"].(map[string]any); ok {
			if errMsg, ok := errObj["message"].(string); ok {
				msg += ": " + errMsg
			}
		}
		return nil, fmt.Errorf("transport: %w", &ServerError{Message: msg, StatusCode: resp.StatusCode})
	}

	return result, nil
}

// Get sends a GET request and returns the raw response body.
func (c *Client) Get(ctx context.Context, url string, extraHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("transport: create GET request: %w", err)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	req.AddCookie(&http.Cookie{Name: "SOCS", Value: "CAI"})

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("transport: execute GET request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("transport: read GET response: %w", err)
	}
	return data, nil
}

// ServerError is returned when the API responds with an HTTP error status.
type ServerError struct {
	Message    string
	StatusCode int
}

func (e *ServerError) Error() string { return e.Message }
