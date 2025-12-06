package errs

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseConnection = errors.New("database connection failed")
)

// Database & Storage Specific Errors
var (
	ErrPoolExhausted             = errors.New("connection pool exhausted")
	ErrDeadlock                  = errors.New("database deadlock")
	ErrSerializationFailure      = errors.New("serialization failure")
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")
	ErrReplicaLag                = errors.New("replica lag")
	ErrMigrationMismatch         = errors.New("migration mismatch")
	ErrStorageQuotaFull          = errors.New("storage quota full")
	ErrTransactionFailed         = errors.New("transaction failed")
	ErrForeignKeyConstraint      = errors.New("foreign key constraint violation")
	ErrDatabaseTimeout           = errors.New("database timeout")
	ErrDatabaseLock              = errors.New("database lock timeout")
	ErrDatabaseCorruption        = errors.New("database corruption")
)

func NewAlreadyExists(entity string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusConflict,
		err:        fmt.Errorf("%s %w", entity, ErrAlreadyExists),
	}
}

func NewNotFound(entity string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusNotFound,
		err:        fmt.Errorf("%s %w", entity, ErrNotFound),
	}
}

// NewDatabaseError creates a new database error with details about the operation
func NewDatabaseError(operation, entity string, cause error) *ApiErr {
	details := fmt.Sprintf("Failed to %s %s", operation, entity)

	// Check for common database errors and provide more specific messages
	if cause != nil {
		errStr := cause.Error()
		switch {
		case strings.Contains(errStr, "duplicate key"):
			return &ApiErr{
				StatusCode: http.StatusConflict,
				err:        fmt.Errorf("%s already exists", entity),
				Details:    details,
				Cause:      cause,
			}
		case strings.Contains(errStr, "foreign key constraint"):
			return &ApiErr{
				StatusCode: http.StatusBadRequest,
				err:        fmt.Errorf("invalid reference in %s", entity),
				Details:    "The referenced resource does not exist or cannot be linked",
				Cause:      cause,
			}
		case strings.Contains(errStr, "not found"):
			return &ApiErr{
				StatusCode: http.StatusNotFound,
				err:        fmt.Errorf("%s not found", entity),
				Details:    details,
				Cause:      cause,
			}
		case strings.Contains(errStr, "connection"):
			return &ApiErr{
				StatusCode: http.StatusServiceUnavailable,
				err:        ErrDatabaseConnection,
				Details:    "Unable to connect to database",
				Cause:      cause,
			}
		}
	}

	// Generic database error
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrDatabaseQuery,
		Details:    details,
		Cause:      cause,
	}
}

// Database & Storage Error Constructors
func NewPoolExhaustedError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrPoolExhausted,
		Details:    fmt.Sprintf("Connection pool exhausted during %s", operation),
		Field:      "connection_pool",
	}
}

func NewDeadlockError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusConflict,
		err:        ErrDeadlock,
		Details:    fmt.Sprintf("Database deadlock during %s", operation),
		Cause:      cause,
		Field:      "deadlock",
	}
}

func NewSerializationFailureError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusConflict,
		err:        ErrSerializationFailure,
		Details:    fmt.Sprintf("Serialization failure during %s", operation),
		Cause:      cause,
		Field:      "serialization",
	}
}

func NewUniqueConstraintViolationError(entity, field string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusConflict,
		err:        ErrUniqueConstraintViolation,
		Details:    fmt.Sprintf("Unique constraint violation on %s.%s", entity, field),
		Cause:      cause,
		Field:      field,
	}
}

func NewReplicaLagError(operation string, lag time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrReplicaLag,
		Details:    fmt.Sprintf("Replica lag detected during %s: %v", operation, lag),
		Field:      "replica_lag",
	}
}

func NewMigrationMismatchError(expected, actual string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrMigrationMismatch,
		Details:    fmt.Sprintf("Migration mismatch: expected %s, got %s", expected, actual),
		Field:      "migration",
	}
}

func NewStorageQuotaFullError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInsufficientStorage,
		err:        ErrStorageQuotaFull,
		Details:    fmt.Sprintf("Storage quota full during %s", operation),
		Field:      "storage",
	}
}

func NewTransactionFailedError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrTransactionFailed,
		Details:    fmt.Sprintf("Transaction failed during %s", operation),
		Cause:      cause,
		Field:      "transaction",
	}
}

func NewForeignKeyConstraintError(entity, referencedEntity string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrForeignKeyConstraint,
		Details:    fmt.Sprintf("Foreign key constraint violation: %s references %s", entity, referencedEntity),
		Cause:      cause,
		Field:      "foreign_key",
	}
}

func NewDatabaseTimeoutError(operation string, timeout time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusRequestTimeout,
		err:        ErrDatabaseTimeout,
		Details:    fmt.Sprintf("Database timeout during %s after %v", operation, timeout),
		Field:      "timeout",
	}
}

func NewDatabaseLockError(operation string, lockType string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusConflict,
		err:        ErrDatabaseLock,
		Details:    fmt.Sprintf("Database lock timeout during %s (lock type: %s)", operation, lockType),
		Field:      "lock",
	}
}

func NewDatabaseCorruptionError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrDatabaseCorruption,
		Details:    fmt.Sprintf("Database corruption detected during %s", operation),
		Cause:      cause,
		Field:      "corruption",
	}
}

// Database & Storage Error Type Checkers
func IsPoolExhaustedError(err error) bool {
	return errors.Is(err, ErrPoolExhausted)
}

func IsDeadlockError(err error) bool {
	return errors.Is(err, ErrDeadlock)
}

func IsSerializationFailureError(err error) bool {
	return errors.Is(err, ErrSerializationFailure)
}

func IsUniqueConstraintViolationError(err error) bool {
	return errors.Is(err, ErrUniqueConstraintViolation)
}

func IsReplicaLagError(err error) bool {
	return errors.Is(err, ErrReplicaLag)
}

func IsMigrationMismatchError(err error) bool {
	return errors.Is(err, ErrMigrationMismatch)
}

func IsStorageQuotaFullError(err error) bool {
	return errors.Is(err, ErrStorageQuotaFull)
}

func IsTransactionFailedError(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}

func IsForeignKeyConstraintError(err error) bool {
	return errors.Is(err, ErrForeignKeyConstraint)
}

func IsDatabaseTimeoutError(err error) bool {
	return errors.Is(err, ErrDatabaseTimeout)
}

func IsDatabaseLockError(err error) bool {
	return errors.Is(err, ErrDatabaseLock)
}

func IsDatabaseCorruptionError(err error) bool {
	return errors.Is(err, ErrDatabaseCorruption)
}
