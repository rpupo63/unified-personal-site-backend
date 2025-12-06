package errs

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	Unauthorized = NewApiErr(http.StatusUnauthorized, "unauthorized")
)

// Authentication & Authorization Errors
var (
	ErrMissingToken      = errors.New("missing access token")
	ErrExpiredToken      = errors.New("expired access token")
	ErrInvalidToken      = errors.New("invalid access token")
	ErrInsufficientScope = errors.New("insufficient scope")
	ErrInsufficientRole  = errors.New("insufficient role")
	ErrTokenExpired      = errors.New("token expired")
)

// Concurrency & Synchronization Errors
var (
	ErrGoroutineLeak     = errors.New("goroutine leak")
	ErrDataRace          = errors.New("data race detected")
	ErrStarvation        = errors.New("starvation detected")
	ErrPriorityInversion = errors.New("priority inversion")
)

func Malformed(payloadName string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, payloadName+" malformed")
}

func BadRequest(message string) *ApiErr {
	return NewApiErr(http.StatusBadRequest, message)
}

// Authentication & Authorization Error Constructors
func NewMissingTokenError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusUnauthorized,
		err:        ErrMissingToken,
		Details:    "Missing access token",
		Field:      "authorization",
	}
}

func NewExpiredTokenError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusUnauthorized,
		err:        ErrExpiredToken,
		Details:    "Access token has expired",
		Field:      "authorization",
	}
}

func NewInvalidTokenError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusUnauthorized,
		err:        ErrInvalidToken,
		Details:    "Invalid access token",
		Field:      "authorization",
	}
}

func NewInsufficientScopeError(requiredScope string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusForbidden,
		err:        ErrInsufficientScope,
		Details:    fmt.Sprintf("Insufficient scope. Required: %s", requiredScope),
		Field:      "authorization",
	}
}

func NewInsufficientRoleError(requiredRole string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusForbidden,
		err:        ErrInsufficientRole,
		Details:    fmt.Sprintf("Insufficient role. Required: %s", requiredRole),
		Field:      "authorization",
	}
}

func NewTokenExpiredError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusUnauthorized,
		err:        ErrTokenExpired,
		Details:    "Token has expired",
		Field:      "authorization",
	}
}

// Concurrency & Synchronization Error Constructors
func NewGoroutineLeakError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrGoroutineLeak,
		Details:    fmt.Sprintf("Goroutine leak detected in %s", operation),
		Field:      "concurrency",
	}
}

func NewDataRaceError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrDataRace,
		Details:    fmt.Sprintf("Data race detected in %s", operation),
		Field:      "concurrency",
	}
}

func NewStarvationError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrStarvation,
		Details:    fmt.Sprintf("Starvation detected in %s", operation),
		Field:      "concurrency",
	}
}

func NewPriorityInversionError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrPriorityInversion,
		Details:    fmt.Sprintf("Priority inversion detected in %s", operation),
		Field:      "concurrency",
	}
}

// Authentication & Authorization Error Type Checkers
func IsMissingTokenError(err error) bool {
	return errors.Is(err, ErrMissingToken)
}

func IsExpiredTokenError(err error) bool {
	return errors.Is(err, ErrExpiredToken)
}

func IsInvalidTokenError(err error) bool {
	return errors.Is(err, ErrInvalidToken)
}

func IsInsufficientScopeError(err error) bool {
	return errors.Is(err, ErrInsufficientScope)
}

func IsInsufficientRoleError(err error) bool {
	return errors.Is(err, ErrInsufficientRole)
}

func IsTokenExpiredError(err error) bool {
	return errors.Is(err, ErrTokenExpired)
}

// Concurrency & Synchronization Error Type Checkers
func IsGoroutineLeakError(err error) bool {
	return errors.Is(err, ErrGoroutineLeak)
}

func IsDataRaceError(err error) bool {
	return errors.Is(err, ErrDataRace)
}

func IsStarvationError(err error) bool {
	return errors.Is(err, ErrStarvation)
}

func IsPriorityInversionError(err error) bool {
	return errors.Is(err, ErrPriorityInversion)
}
