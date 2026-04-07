package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	maxRetries     = 3
	initialBackoff = 500 * time.Millisecond
)

type Client struct {
	baseURL    string
	apiKey     string
	version    string
	httpClient *http.Client
}

type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

func NewClient(baseURL, apiKey, version string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		version: version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) DoRequest(method, path string, body interface{}) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(initialBackoff * time.Duration(1<<(attempt-1)))
		}

		var bodyReader io.Reader
		if body != nil {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", "terraform-provider-manako/"+c.version)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
					time.Sleep(time.Duration(seconds) * time.Second)
				}
			}
			lastErr = &APIError{StatusCode: 429, Code: "RATE_LIMIT", Message: "Too many requests"}
			continue
		}

		if resp.StatusCode >= 400 {
			apiErr := &APIError{StatusCode: resp.StatusCode}
			var errResp struct {
				Error struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if json.Unmarshal(respBody, &errResp) == nil {
				apiErr.Code = errResp.Error.Code
				apiErr.Message = errResp.Error.Message
			}
			return nil, apiErr
		}

		return respBody, nil
	}

	return nil, lastErr
}

// Monitor types

type Monitor struct {
	ID                     string                 `json:"id"`
	TeamID                 string                 `json:"teamId"`
	Type                   string                 `json:"type"`
	Name                   string                 `json:"name"`
	Config                 map[string]interface{} `json:"config"`
	IntervalSeconds        int64                  `json:"intervalSeconds"`
	IsActive               int                    `json:"isActive"`
	ServiceID              string                 `json:"serviceId"`
	Status                 string                 `json:"status"`
	LastCheckedAt          *string                `json:"lastCheckedAt"`
	LastErrorMessage       *string                `json:"lastErrorMessage"`
	MaintenanceUntil       *string                `json:"maintenanceUntil"`
	AutoResolveMaintenance int                    `json:"autoResolveMaintenance"`
	CreatedAt              string                 `json:"createdAt"`
	UpdatedAt              string                 `json:"updatedAt"`
}

type CreateMonitorRequest struct {
	Type            string                 `json:"type"`
	Name            string                 `json:"name,omitempty"`
	Config          map[string]interface{} `json:"config"`
	IntervalSeconds int64                  `json:"intervalSeconds,omitempty"`
	ServiceID       string                 `json:"serviceId,omitempty"`
}

type UpdateMonitorRequest struct {
	Name            string                 `json:"name,omitempty"`
	Config          map[string]interface{} `json:"config,omitempty"`
	IntervalSeconds *int64                 `json:"intervalSeconds,omitempty"`
	IsActive        *bool                  `json:"isActive,omitempty"`
}

// Monitor CRUD

func (c *Client) GetMonitor(id string) (*Monitor, error) {
	data, err := c.DoRequest("GET", "/api/v1/monitors/"+id, nil)
	if err != nil {
		return nil, err
	}
	var m Monitor
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse monitor response: %w", err)
	}
	return &m, nil
}

func (c *Client) ListMonitors() ([]Monitor, error) {
	data, err := c.DoRequest("GET", "/api/v1/monitors", nil)
	if err != nil {
		return nil, err
	}
	var monitors []Monitor
	if err := json.Unmarshal(data, &monitors); err != nil {
		return nil, fmt.Errorf("failed to parse monitors response: %w", err)
	}
	return monitors, nil
}

func (c *Client) CreateMonitor(req CreateMonitorRequest) (*Monitor, error) {
	data, err := c.DoRequest("POST", "/api/v1/monitors", req)
	if err != nil {
		return nil, err
	}
	var m Monitor
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse monitor response: %w", err)
	}
	return &m, nil
}

func (c *Client) UpdateMonitor(id string, req UpdateMonitorRequest) (*Monitor, error) {
	data, err := c.DoRequest("PUT", "/api/v1/monitors/"+id, req)
	if err != nil {
		return nil, err
	}
	var m Monitor
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse monitor response: %w", err)
	}
	return &m, nil
}

func (c *Client) DeleteMonitor(id string) error {
	_, err := c.DoRequest("DELETE", "/api/v1/monitors/"+id, nil)
	return err
}
