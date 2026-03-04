package client

import (
	"errors"
	"fmt"
	"net/http"
)

// errorResponse is the top-level error response envelope from the API.
type errorResponse struct {
	Type      string    `json:"type"`
	Error     errorBody `json:"error"`
	RequestID string    `json:"request_id"`
}

type errorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// APIError represents an error response from the Claude Admin API.
type APIError struct {
	StatusCode int    `json:"-"`
	Type       string `json:"type"`
	Message    string `json:"message"`
	RequestID  string `json:"-"`
}

func (e *APIError) Error() string {
	typ := ""
	if e.Type != "" {
		typ = ", " + e.Type
	}
	reqID := ""
	if e.RequestID != "" {
		reqID = " (request_id: " + e.RequestID + ")"
	}
	return fmt.Sprintf("API error (HTTP %d%s): %s%s", e.StatusCode, typ, e.Message, reqID)
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
