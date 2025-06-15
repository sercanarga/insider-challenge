package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"insider-challenge/pkg/config"
	"insider-challenge/pkg/errors"
)

// WebhookResponse represents the response from the webhook
type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

// HTTPClient handles http operations for the service
type HTTPClient struct {
	cfg            *config.Config
	httpTimeout    time.Duration
	requestTimeout time.Duration
}

// NewHTTPClient creates a new http client instance
func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		cfg:            cfg,
		httpTimeout:    10 * time.Second,
		requestTimeout: 5 * time.Second,
	}
}

// SendRequest handles the http request to the webhook
func (c *HTTPClient) SendRequest(ctx context.Context, jsonData []byte) (WebhookResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return WebhookResponse{}, errors.Wrap(err, "create request")
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: c.httpTimeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: c.requestTimeout,
			ExpectContinueTimeout: c.requestTimeout,
			IdleConnTimeout:       c.requestTimeout,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return WebhookResponse{}, errors.Wrap(errors.ErrWebhookFailed, "request timeout exceeded")
		}
		return WebhookResponse{}, errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return WebhookResponse{}, errors.Wrap(errors.ErrWebhookFailed, fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
	}

	var response WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return WebhookResponse{}, errors.Wrap(err, "decode response")
	}

	return response, nil
}
