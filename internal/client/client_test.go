package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.example.com", "mk_testapikey123", "0.1.0")
	if c.baseURL != "https://api.example.com" {
		t.Errorf("expected baseURL https://api.example.com, got %s", c.baseURL)
	}
	if c.apiKey != "mk_testapikey123" {
		t.Errorf("expected apiKey mk_testapikey123, got %s", c.apiKey)
	}
}

func TestClient_DoRequest_SetsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer mk_test123" {
			t.Errorf("expected Authorization Bearer mk_test123, got %s", got)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("expected User-Agent header to be set")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "mk_test123", "0.1.0")
	_, err := c.DoRequest("GET", "/api/v1/monitors", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_DoRequest_404ReturnsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"NOT_FOUND","message":"Monitor not found","status":404}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "mk_test123", "0.1.0")
	_, err := c.DoRequest("GET", "/api/v1/monitors/abc", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if !apiErr.IsNotFound() {
		t.Error("expected IsNotFound() to return true")
	}
}

func TestClient_DoRequest_429Retries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT","message":"Too many requests","status":429}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"01ABC"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "mk_test123", "0.1.0")
	resp, err := c.DoRequest("GET", "/api/v1/monitors/01ABC", nil)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
