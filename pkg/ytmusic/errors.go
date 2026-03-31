// Package ytmusic provides a Go client for the YouTube Music API.
package ytmusic

import (
	"errors"
	"fmt"
)

// ErrServer is returned when the YouTube Music server returns a non-2xx status.
var ErrServer = errors.New("ytmusic: server error")

// ErrUser is returned for invalid user input.
var ErrUser = errors.New("ytmusic: user error")

// ServerError wraps a server-returned error with the HTTP status code.
type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("ytmusic: server error %d: %s", e.StatusCode, e.Message)
}
func (e *ServerError) Is(target error) bool { return target == ErrServer }

// UserError represents invalid caller input.
type UserError struct {
	Message string
}

func (e *UserError) Error() string        { return "ytmusic: " + e.Message }
func (e *UserError) Is(target error) bool { return target == ErrUser }
