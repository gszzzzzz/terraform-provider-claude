package client

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error response from the Claude Admin API.
type APIError struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Type != "" {
		return fmt.Sprintf("API error (HTTP %d, %s): %s", e.StatusCode, e.Type, e.Message)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsAlreadyArchived returns true if the error indicates the resource is already archived.
func IsAlreadyArchived(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusBadRequest && apiErr.Type == "already_archived"
	}
	return false
}
