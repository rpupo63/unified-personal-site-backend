package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Common error sentinel values
var (
	ErrForbidden    = errors.New("operation not allowed")
	ErrBadRequest   = errors.New("malformed request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInternal     = errors.New("internal server error")
	ErrConflict     = errors.New("resource conflict")
	ErrCORSBlocked  = errors.New("request blocked by CORS policy")
)

// Request & Input-Validation Errors
var (
	ErrMalformedPayload     = errors.New("malformed payload")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidField         = errors.New("invalid field")
	ErrUnsupportedMediaType = errors.New("unsupported media type")
	ErrMaxBodySizeExceeded  = errors.New("max body size exceeded")
	ErrInvalidJSON          = errors.New("invalid JSON")
	ErrInvalidXML           = errors.New("invalid XML")
	ErrInvalidContentType   = errors.New("invalid content type")
	ErrInvalidCharset       = errors.New("invalid charset")
)

type ApiErr struct {
	StatusCode int
	err        error
	Details    string // Additional details about the error
	Field      string // Field that caused the error (for validation errors)
	Cause      error  // The underlying cause of the error
}

func NewApiErr(statusCode int, message string) *ApiErr {
	return &ApiErr{
		StatusCode: statusCode,
		err:        errors.New(message),
	}
}

// implements error interface. this allows us to pass an instance of ApiErr as an argument of type `error`
func (e *ApiErr) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.err.Error(), e.Details)
	}
	return e.err.Error()
}

// GetFullError returns a recursive error message including all causes
func (e *ApiErr) GetFullError() string {
	msg := e.Error()
	if e.Cause != nil {
		// Check if the cause is also an ApiErr for recursive error handling
		if apiErr, ok := e.Cause.(*ApiErr); ok {
			msg = fmt.Sprintf("%s -> %s", msg, apiErr.GetFullError())
		} else {
			msg = fmt.Sprintf("%s -> %s", msg, e.Cause.Error())
		}
	}
	return msg
}

// this function allows us to do the following:
// err := &ApiErr{StatusCode: ..., err: someSentinelError}
// errors.Is(err, someSentinelError) ==> evaluates to true
func (e *ApiErr) Unwrap() error {
	return e.err
}

// Common error constructors with appropriate HTTP status codes
func NewNotFoundError(message string) *ApiErr {
	return &ApiErr{StatusCode: 404, err: errors.New(message)}
}

func NewForbiddenError(message string) *ApiErr {
	return &ApiErr{StatusCode: 403, err: errors.New(message)}
}

func NewBadRequestError(message string) *ApiErr {
	return &ApiErr{StatusCode: 400, err: errors.New(message)}
}

func NewUnauthorizedError(message string) *ApiErr {
	return &ApiErr{StatusCode: 401, err: errors.New(message)}
}

func NewInternalError(message string) *ApiErr {
	return &ApiErr{StatusCode: 500, err: errors.New(message)}
}

func NewConflictError(message string) *ApiErr {
	return &ApiErr{StatusCode: 409, err: errors.New(message)}
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// New detailed error constructors
func NewBadRequestErrorWithDetails(message, details string) *ApiErr {
	return &ApiErr{
		StatusCode: 400,
		err:        errors.New(message),
		Details:    details,
	}
}

func NewBadRequestErrorWithField(message, field, details string) *ApiErr {
	return &ApiErr{
		StatusCode: 400,
		err:        errors.New(message),
		Field:      field,
		Details:    details,
	}
}

func NewInternalErrorWithCause(message string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: 500,
		err:        errors.New(message),
		Cause:      cause,
	}
}

func NewCORSError(origin string) *ApiErr {
	return &ApiErr{
		StatusCode: 403,
		err:        ErrCORSBlocked,
		Details:    fmt.Sprintf("Origin '%s' is not allowed by CORS policy", origin),
	}
}

// Request & Input-Validation Error Constructors
func NewMalformedPayloadError(payloadType string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrMalformedPayload,
		Details:    fmt.Sprintf("Malformed %s payload", payloadType),
		Cause:      cause,
		Field:      "payload",
	}
}

func NewMissingRequiredFieldError(fieldName string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrMissingRequiredField,
		Details:    fmt.Sprintf("Missing required field: %s", fieldName),
		Field:      fieldName,
	}
}

func NewInvalidFieldError(fieldName string, reason string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrInvalidField,
		Details:    fmt.Sprintf("Invalid field %s: %s", fieldName, reason),
		Field:      fieldName,
	}
}

func NewUnsupportedMediaTypeError(contentType string, allowedTypes []string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusUnsupportedMediaType,
		err:        ErrUnsupportedMediaType,
		Details:    fmt.Sprintf("Unsupported media type: %s. Allowed types: %v", contentType, allowedTypes),
		Field:      "content_type",
	}
}

func NewMaxBodySizeExceededError(maxSize int64) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusRequestEntityTooLarge,
		err:        ErrMaxBodySizeExceeded,
		Details:    fmt.Sprintf("Request body size exceeded maximum allowed size of %d bytes", maxSize),
		Field:      "body_size",
	}
}

func NewInvalidJSONError(cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrInvalidJSON,
		Details:    "Invalid JSON format",
		Cause:      cause,
		Field:      "json",
	}
}

func NewInvalidXMLError(cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrInvalidXML,
		Details:    "Invalid XML format",
		Cause:      cause,
		Field:      "xml",
	}
}

func NewInvalidContentTypeError(contentType string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrInvalidContentType,
		Details:    fmt.Sprintf("Invalid content type: %s", contentType),
		Field:      "content_type",
	}
}

func NewInvalidCharsetError(charset string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrInvalidCharset,
		Details:    fmt.Sprintf("Invalid charset: %s", charset),
		Field:      "charset",
	}
}

// Request & Input-Validation Error Type Checkers
func IsMalformedPayloadError(err error) bool {
	return errors.Is(err, ErrMalformedPayload)
}

func IsMissingRequiredFieldError(err error) bool {
	return errors.Is(err, ErrMissingRequiredField)
}

func IsInvalidFieldError(err error) bool {
	return errors.Is(err, ErrInvalidField)
}

func IsUnsupportedMediaTypeError(err error) bool {
	return errors.Is(err, ErrUnsupportedMediaType)
}

func IsMaxBodySizeExceededError(err error) bool {
	return errors.Is(err, ErrMaxBodySizeExceeded)
}

func IsInvalidJSONError(err error) bool {
	return errors.Is(err, ErrInvalidJSON)
}

func IsInvalidXMLError(err error) bool {
	return errors.Is(err, ErrInvalidXML)
}

func IsInvalidContentTypeError(err error) bool {
	return errors.Is(err, ErrInvalidContentType)
}

func IsInvalidCharsetError(err error) bool {
	return errors.Is(err, ErrInvalidCharset)
}
