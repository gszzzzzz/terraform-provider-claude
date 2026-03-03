package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		expectedBaseURL string
	}{
		{
			name:            "empty baseURL uses default",
			baseURL:         "",
			expectedBaseURL: defaultBaseURL,
		},
		{
			name:            "custom baseURL preserved",
			baseURL:         "https://custom.example.com",
			expectedBaseURL: "https://custom.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewClient("test-key", tt.baseURL)
			if c.BaseURL != tt.expectedBaseURL {
				t.Errorf("BaseURL = %q, want %q", c.BaseURL, tt.expectedBaseURL)
			}
			if c.HTTPClient == nil {
				t.Error("HTTPClient is nil")
			}
			if c.UserAgent != "terraform-provider-claude" {
				t.Errorf("UserAgent = %q, want %q", c.UserAgent, "terraform-provider-claude")
			}
			if c.APIKey != "test-key" {
				t.Errorf("APIKey = %q, want %q", c.APIKey, "test-key")
			}
			if c.HTTPClient.Timeout != 30*time.Second {
				t.Errorf("HTTPClient.Timeout = %v, want %v", c.HTTPClient.Timeout, 30*time.Second)
			}
		})
	}
}

func TestDoRequest_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "test-key" {
			t.Errorf("x-api-key = %q, want %q", got, "test-key")
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", got, "2023-06-01")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want %q", got, "application/json")
		}
		if got := r.Header.Get("User-Agent"); got != "terraform-provider-claude" {
			t.Errorf("User-Agent = %q, want %q", got, "terraform-provider-claude")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient("test-key", server.URL)
	err := c.doRequest(context.Background(), http.MethodGet, "/test", nil, nil)
	if err != nil {
		t.Fatalf("doRequest returned error: %v", err)
	}
}

func TestDoRequest_SuccessResponse(t *testing.T) {
	type testStruct struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name        string
		statusCode  int
		body        string
		result      any
		checkResult func(t *testing.T, result any)
	}{
		{
			name:       "200 with JSON",
			statusCode: 200,
			body:       `{"id":"1","name":"test"}`,
			result:     &testStruct{},
			checkResult: func(t *testing.T, result any) {
				ts, ok := result.(*testStruct)
				if !ok {
					t.Fatal("expected *testStruct")
				}
				if ts.ID != "1" || ts.Name != "test" {
					t.Errorf("got %+v, want {ID:1 Name:test}", ts)
				}
			},
		},
		{
			name:       "201 success",
			statusCode: 201,
			body:       `{"id":"1","name":"created"}`,
			result:     &testStruct{},
			checkResult: func(t *testing.T, result any) {
				ts, ok := result.(*testStruct)
				if !ok {
					t.Fatal("expected *testStruct")
				}
				if ts.ID != "1" {
					t.Errorf("got ID=%q, want %q", ts.ID, "1")
				}
			},
		},
		{
			name:        "result nil",
			statusCode:  200,
			body:        `{}`,
			result:      nil,
			checkResult: nil,
		},
		{
			name:        "empty body with nil result",
			statusCode:  204,
			body:        "",
			result:      nil,
			checkResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.body != "" {
					_, _ = w.Write([]byte(tt.body))
				}
			}))
			defer server.Close()

			c := NewClient("test-key", server.URL)
			err := c.doRequest(context.Background(), http.MethodGet, "/test", nil, tt.result)
			if err != nil {
				t.Fatalf("doRequest returned error: %v", err)
			}
			if tt.checkResult != nil {
				tt.checkResult(t, tt.result)
			}
		})
	}
}

func TestDoRequest_ErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		checkError func(t *testing.T, err error)
	}{
		{
			name:       "400 JSON error",
			statusCode: 400,
			body:       `{"type":"invalid_request","message":"bad"}`,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Fatalf("expected *APIError, got %T", err)
				}
				if apiErr.StatusCode != 400 {
					t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
				}
				if apiErr.Type != "invalid_request" {
					t.Errorf("Type = %q, want %q", apiErr.Type, "invalid_request")
				}
				if apiErr.Message != "bad" {
					t.Errorf("Message = %q, want %q", apiErr.Message, "bad")
				}
			},
		},
		{
			name:       "404 not found",
			statusCode: 404,
			body:       `{"type":"not_found","message":"gone"}`,
			checkError: func(t *testing.T, err error) {
				if !IsNotFound(err) {
					t.Error("expected IsNotFound to be true")
				}
			},
		},
		{
			name:       "500 non-JSON",
			statusCode: 500,
			body:       "Internal Server Error",
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Fatalf("expected *APIError, got %T", err)
				}
				if apiErr.Message != "Internal Server Error" {
					t.Errorf("Message = %q, want %q", apiErr.Message, "Internal Server Error")
				}
			},
		},
		{
			name:       "401 auth error",
			statusCode: 401,
			body:       `{"type":"auth_error","message":"key"}`,
			checkError: func(t *testing.T, err error) {
				apiErr, ok := err.(*APIError)
				if !ok {
					t.Fatalf("expected *APIError, got %T", err)
				}
				if apiErr.StatusCode != 401 {
					t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
				}
				if apiErr.Type != "auth_error" {
					t.Errorf("Type = %q, want %q", apiErr.Type, "auth_error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c := NewClient("test-key", server.URL)
			err := c.doRequest(context.Background(), http.MethodGet, "/test", nil, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			tt.checkError(t, err)
		})
	}
}

func TestDoRequest_RequestBody(t *testing.T) {
	t.Run("struct body", func(t *testing.T) {
		type reqBody struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading request body: %v", err)
			}
			var got reqBody
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshaling request body: %v", err)
			}
			if got.Name != "test" || got.Age != 30 {
				t.Errorf("got %+v, want {Name:test Age:30}", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient("test-key", server.URL)
		err := c.doRequest(context.Background(), http.MethodPost, "/test", reqBody{Name: "test", Age: 30}, nil)
		if err != nil {
			t.Fatalf("doRequest returned error: %v", err)
		}
	})

	t.Run("nil body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("reading request body: %v", err)
			}
			if len(body) != 0 {
				t.Errorf("expected empty body, got %q", string(body))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		c := NewClient("test-key", server.URL)
		err := c.doRequest(context.Background(), http.MethodGet, "/test", nil, nil)
		if err != nil {
			t.Fatalf("doRequest returned error: %v", err)
		}
	})
}
