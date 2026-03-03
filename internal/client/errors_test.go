package client

import (
	"errors"
	"fmt"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with type",
			err:      &APIError{StatusCode: 400, Type: "invalid_request", Message: "bad param"},
			expected: "API error (HTTP 400, invalid_request): bad param",
		},
		{
			name:     "without type",
			err:      &APIError{StatusCode: 500, Type: "", Message: "internal error"},
			expected: "API error (HTTP 500): internal error",
		},
		{
			name:     "empty message",
			err:      &APIError{StatusCode: 401, Type: "unauthorized", Message: ""},
			expected: "API error (HTTP 401, unauthorized): ",
		},
		{
			name:     "both empty",
			err:      &APIError{StatusCode: 503, Type: "", Message: ""},
			expected: "API error (HTTP 503): ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "404 APIError",
			err:      &APIError{StatusCode: 404},
			expected: true,
		},
		{
			name:     "400 APIError",
			err:      &APIError{StatusCode: 400},
			expected: false,
		},
		{
			name:     "wrapped 404",
			err:      fmt.Errorf("wrap: %w", &APIError{StatusCode: 404}),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "plain error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}
