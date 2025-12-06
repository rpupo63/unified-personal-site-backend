package errs

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Third-Party API & LLM Specific Errors
var (
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrModelOverloaded        = errors.New("model overloaded")
	ErrContextLengthExceeded  = errors.New("context length exceeded")
	ErrContentPolicyViolation = errors.New("content policy violation")
	ErrBillingQuotaExhausted  = errors.New("billing quota exhausted")
	ErrStreamingChunkDropped  = errors.New("streaming chunk dropped")
	ErrInvalidAPIKey          = errors.New("invalid API key")
	ErrServiceUnavailable     = errors.New("service unavailable")
	ErrTimeout                = errors.New("timeout")
	ErrCircuitBreakerOpen     = errors.New("circuit breaker open")
)

// Configuration & Environment Errors
var (
	ErrConfigMissing       = errors.New("configuration missing")
	ErrConfigInvalid       = errors.New("configuration invalid")
	ErrRegionNotSupported  = errors.New("region not supported")
	ErrSecretMismatch      = errors.New("secret mismatch")
	ErrEnvironmentVariable = errors.New("environment variable error")
)

// Networking & Transport Errors
var (
	ErrDNSResolution      = errors.New("DNS resolution failed")
	ErrTCPTimeout         = errors.New("TCP connection timeout")
	ErrTLSHandshake       = errors.New("TLS handshake failed")
	ErrConnectionReset    = errors.New("connection reset")
	ErrProxyBlocked       = errors.New("proxy/firewall blocked")
	ErrNetworkUnreachable = errors.New("network unreachable")
)

// Resource Exhaustion Errors
var (
	ErrOutOfMemory         = errors.New("out of memory")
	ErrCPUExhausted        = errors.New("CPU exhausted")
	ErrFileDescriptorLimit = errors.New("file descriptor limit exceeded")
	ErrDiskSpaceFull       = errors.New("disk space full")
)

// Timeout & Cancellation Errors
var (
	ErrContextDeadline    = errors.New("context deadline exceeded")
	ErrClientDisconnected = errors.New("client disconnected")
	ErrComputationTimeout = errors.New("computation timeout")
	ErrRequestTimeout     = errors.New("request timeout")
)

// Data Consistency & Integrity Errors
var (
	ErrSchemaVersionMismatch = errors.New("schema version mismatch")
	ErrClockSkew             = errors.New("clock skew detected")
	ErrPartialFailure        = errors.New("partial failure")
	ErrDataCorruption        = errors.New("data corruption")
)

// Computational & Business Logic Errors
var (
	ErrDivideByZero = errors.New("divide by zero")
	ErrOverflow     = errors.New("arithmetic overflow")
	ErrUnderflow    = errors.New("arithmetic underflow")
	ErrNilPointer   = errors.New("nil pointer dereference")
	ErrInvalidInput = errors.New("invalid input")
)

// Serialization & Encoding Errors
var (
	ErrCircularStructure = errors.New("circular structure detected")
	ErrBase64Decode      = errors.New("base64 decode error")
	ErrCharsetConversion = errors.New("charset conversion error")
	ErrJSONMarshal       = errors.New("JSON marshal error")
	ErrJSONUnmarshal     = errors.New("JSON unmarshal error")
)

// Dependency & Service Discovery Errors
var (
	ErrServiceUnreachable = errors.New("service unreachable")
	ErrServiceDiscovery   = errors.New("service discovery failed")
	ErrFeatureFlagBackend = errors.New("feature flag backend unavailable")
	ErrLoadBalancer       = errors.New("load balancer error")
)

// Security & Compliance Errors
var (
	ErrSQLInjection       = errors.New("SQL injection attempt")
	ErrRequestForgery     = errors.New("request forgery detected")
	ErrSecretsLeaked      = errors.New("secrets leaked")
	ErrPIIBreach          = errors.New("PII breach detected")
	ErrUnauthorizedAccess = errors.New("unauthorized access")
)

// Observability & Telemetry Errors
var (
	ErrTracerExporter     = errors.New("tracer exporter down")
	ErrLogPipeline        = errors.New("log pipeline error")
	ErrMetricsCardinality = errors.New("metrics cardinality explosion")
	ErrTelemetryFailure   = errors.New("telemetry failure")
)

// Deployment & Runtime Platform Errors
var (
	ErrBadRollout     = errors.New("bad rollout")
	ErrSidecarCrash   = errors.New("sidecar crash")
	ErrContainerImage = errors.New("container image error")
	ErrPodUnhealthy   = errors.New("pod unhealthy")
)

// LLM Service Specific Error Constructors
func NewRateLimitError(service string, retryAfter time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusTooManyRequests,
		err:        ErrRateLimitExceeded,
		Details:    fmt.Sprintf("Rate limit exceeded for %s service", service),
		Field:      "rate_limit",
	}
}

func NewModelOverloadedError(service string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrModelOverloaded,
		Details:    fmt.Sprintf("Model overloaded for %s service", service),
		Field:      "model_capacity",
	}
}

func NewContextLengthError(service string, maxTokens int) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrContextLengthExceeded,
		Details:    fmt.Sprintf("Context length exceeded for %s service (max: %d tokens)", service, maxTokens),
		Field:      "context_length",
	}
}

func NewContentPolicyError(service string, violation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrContentPolicyViolation,
		Details:    fmt.Sprintf("Content policy violation in %s service: %s", service, violation),
		Field:      "content_policy",
	}
}

func NewBillingQuotaError(service string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusPaymentRequired,
		err:        ErrBillingQuotaExhausted,
		Details:    fmt.Sprintf("Billing quota exhausted for %s service", service),
		Field:      "billing",
	}
}

func NewStreamingError(service string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrStreamingChunkDropped,
		Details:    fmt.Sprintf("Streaming chunk dropped for %s service", service),
		Field:      "streaming",
	}
}

// Configuration & Environment Error Constructors
func NewConfigError(configName string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrConfigMissing,
		Details:    fmt.Sprintf("Configuration error for %s", configName),
		Cause:      cause,
	}
}

func NewEnvironmentVariableError(varName string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrEnvironmentVariable,
		Details:    fmt.Sprintf("Environment variable %s is not set or invalid", varName),
		Field:      varName,
	}
}

func NewRegionNotSupportedError(region string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrRegionNotSupported,
		Details:    fmt.Sprintf("Region %s is not supported", region),
		Field:      "region",
	}
}

// Networking & Transport Error Constructors
func NewDNSError(host string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrDNSResolution,
		Details:    fmt.Sprintf("DNS resolution failed for %s", host),
		Cause:      cause,
	}
}

func NewTCPTimeoutError(host string, timeout time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrTCPTimeout,
		Details:    fmt.Sprintf("TCP connection timeout to %s after %v", host, timeout),
		Field:      "connection",
	}
}

func NewTLSHandshakeError(host string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrTLSHandshake,
		Details:    fmt.Sprintf("TLS handshake failed for %s", host),
		Cause:      cause,
	}
}

func NewConnectionResetError(host string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrConnectionReset,
		Details:    fmt.Sprintf("Connection reset by peer for %s", host),
		Field:      "connection",
	}
}

// Resource Exhaustion Error Constructors
func NewOutOfMemoryError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrOutOfMemory,
		Details:    fmt.Sprintf("Out of memory during %s operation", operation),
		Field:      "memory",
	}
}

func NewCPUExhaustedError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrCPUExhausted,
		Details:    fmt.Sprintf("CPU exhausted during %s operation", operation),
		Field:      "cpu",
	}
}

func NewFileDescriptorLimitError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrFileDescriptorLimit,
		Details:    "File descriptor limit exceeded",
		Field:      "file_descriptors",
	}
}

// Timeout & Cancellation Error Constructors
func NewContextDeadlineError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusRequestTimeout,
		err:        ErrContextDeadline,
		Details:    fmt.Sprintf("Context deadline exceeded for %s", operation),
		Field:      "timeout",
	}
}

func NewClientDisconnectedError() *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusRequestTimeout,
		err:        ErrClientDisconnected,
		Details:    "Client disconnected during request",
		Field:      "client",
	}
}

func NewComputationTimeoutError(operation string, timeout time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusRequestTimeout,
		err:        ErrComputationTimeout,
		Details:    fmt.Sprintf("Computation timeout for %s after %v", operation, timeout),
		Field:      "computation",
	}
}

// Data Consistency & Integrity Error Constructors
func NewSchemaVersionMismatchError(service string, expected, actual string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrSchemaVersionMismatch,
		Details:    fmt.Sprintf("Schema version mismatch in %s service: expected %s, got %s", service, expected, actual),
		Field:      "schema_version",
	}
}

func NewClockSkewError(service string, skew time.Duration) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrClockSkew,
		Details:    fmt.Sprintf("Clock skew detected in %s service: %v", service, skew),
		Field:      "clock",
	}
}

func NewPartialFailureError(operation string, failedSteps []string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrPartialFailure,
		Details:    fmt.Sprintf("Partial failure in %s operation. Failed steps: %v", operation, failedSteps),
		Field:      "partial_failure",
	}
}

// Computational & Business Logic Error Constructors
func NewDivideByZeroError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrDivideByZero,
		Details:    fmt.Sprintf("Divide by zero error in %s", operation),
		Field:      "arithmetic",
	}
}

func NewOverflowError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrOverflow,
		Details:    fmt.Sprintf("Arithmetic overflow in %s", operation),
		Field:      "arithmetic",
	}
}

func NewNilPointerError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrNilPointer,
		Details:    fmt.Sprintf("Nil pointer dereference in %s", operation),
		Field:      "pointer",
	}
}

// Serialization & Encoding Error Constructors
func NewCircularStructureError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrCircularStructure,
		Details:    fmt.Sprintf("Circular structure detected in %s", operation),
		Field:      "serialization",
	}
}

func NewBase64DecodeError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrBase64Decode,
		Details:    fmt.Sprintf("Base64 decode error in %s", operation),
		Cause:      cause,
		Field:      "encoding",
	}
}

func NewJSONMarshalError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrJSONMarshal,
		Details:    fmt.Sprintf("JSON marshal error in %s", operation),
		Cause:      cause,
		Field:      "json",
	}
}

func NewJSONUnmarshalError(operation string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrJSONUnmarshal,
		Details:    fmt.Sprintf("JSON unmarshal error in %s", operation),
		Cause:      cause,
		Field:      "json",
	}
}

// Dependency & Service Discovery Error Constructors
func NewServiceUnreachableError(service string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrServiceUnreachable,
		Details:    fmt.Sprintf("Service %s is unreachable", service),
		Cause:      cause,
		Field:      "service_discovery",
	}
}

func NewServiceDiscoveryError(service string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrServiceDiscovery,
		Details:    fmt.Sprintf("Service discovery failed for %s", service),
		Cause:      cause,
		Field:      "service_discovery",
	}
}

// Security & Compliance Error Constructors
func NewSQLInjectionError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrSQLInjection,
		Details:    fmt.Sprintf("SQL injection attempt detected in %s", operation),
		Field:      "security",
	}
}

func NewRequestForgeryError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusForbidden,
		err:        ErrRequestForgery,
		Details:    fmt.Sprintf("Request forgery detected in %s", operation),
		Field:      "security",
	}
}

func NewSecretsLeakedError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrSecretsLeaked,
		Details:    fmt.Sprintf("Secrets leaked in %s", operation),
		Field:      "security",
	}
}

func NewPIIBreachError(operation string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusBadRequest,
		err:        ErrPIIBreach,
		Details:    fmt.Sprintf("PII breach detected in %s", operation),
		Field:      "compliance",
	}
}

// Observability & Telemetry Error Constructors
func NewTracerExporterError(cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrTracerExporter,
		Details:    "Tracer exporter is down",
		Cause:      cause,
		Field:      "telemetry",
	}
}

func NewLogPipelineError(cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrLogPipeline,
		Details:    "Log pipeline error",
		Cause:      cause,
		Field:      "telemetry",
	}
}

func NewMetricsCardinalityError(metric string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrMetricsCardinality,
		Details:    fmt.Sprintf("Metrics cardinality explosion for %s", metric),
		Field:      "telemetry",
	}
}

// Deployment & Runtime Platform Error Constructors
func NewBadRolloutError(service string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrBadRollout,
		Details:    fmt.Sprintf("Bad rollout for %s service", service),
		Field:      "deployment",
	}
}

func NewSidecarCrashError(service string) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusServiceUnavailable,
		err:        ErrSidecarCrash,
		Details:    fmt.Sprintf("Sidecar crash for %s service", service),
		Field:      "deployment",
	}
}

func NewContainerImageError(service string, cause error) *ApiErr {
	return &ApiErr{
		StatusCode: http.StatusInternalServerError,
		err:        ErrContainerImage,
		Details:    fmt.Sprintf("Container image error for %s service", service),
		Cause:      cause,
		Field:      "deployment",
	}
}

// Error Type Checkers
func IsRateLimitError(err error) bool {
	return errors.Is(err, ErrRateLimitExceeded)
}

func IsModelOverloadedError(err error) bool {
	return errors.Is(err, ErrModelOverloaded)
}

func IsContextLengthError(err error) bool {
	return errors.Is(err, ErrContextLengthExceeded)
}

func IsContentPolicyError(err error) bool {
	return errors.Is(err, ErrContentPolicyViolation)
}

func IsBillingQuotaError(err error) bool {
	return errors.Is(err, ErrBillingQuotaExhausted)
}

func IsStreamingError(err error) bool {
	return errors.Is(err, ErrStreamingChunkDropped)
}

func IsConfigError(err error) bool {
	return errors.Is(err, ErrConfigMissing) || errors.Is(err, ErrConfigInvalid)
}

func IsEnvironmentVariableError(err error) bool {
	return errors.Is(err, ErrEnvironmentVariable)
}

func IsDNSError(err error) bool {
	return errors.Is(err, ErrDNSResolution)
}

func IsTCPTimeoutError(err error) bool {
	return errors.Is(err, ErrTCPTimeout)
}

func IsTLSHandshakeError(err error) bool {
	return errors.Is(err, ErrTLSHandshake)
}

func IsConnectionResetError(err error) bool {
	return errors.Is(err, ErrConnectionReset)
}

func IsOutOfMemoryError(err error) bool {
	return errors.Is(err, ErrOutOfMemory)
}

func IsCPUExhaustedError(err error) bool {
	return errors.Is(err, ErrCPUExhausted)
}

func IsContextDeadlineError(err error) bool {
	return errors.Is(err, ErrContextDeadline)
}

func IsClientDisconnectedError(err error) bool {
	return errors.Is(err, ErrClientDisconnected)
}

func IsComputationTimeoutError(err error) bool {
	return errors.Is(err, ErrComputationTimeout)
}

func IsSchemaVersionMismatchError(err error) bool {
	return errors.Is(err, ErrSchemaVersionMismatch)
}

func IsClockSkewError(err error) bool {
	return errors.Is(err, ErrClockSkew)
}

func IsPartialFailureError(err error) bool {
	return errors.Is(err, ErrPartialFailure)
}

func IsDivideByZeroError(err error) bool {
	return errors.Is(err, ErrDivideByZero)
}

func IsOverflowError(err error) bool {
	return errors.Is(err, ErrOverflow)
}

func IsNilPointerError(err error) bool {
	return errors.Is(err, ErrNilPointer)
}

func IsCircularStructureError(err error) bool {
	return errors.Is(err, ErrCircularStructure)
}

func IsBase64DecodeError(err error) bool {
	return errors.Is(err, ErrBase64Decode)
}

func IsJSONMarshalError(err error) bool {
	return errors.Is(err, ErrJSONMarshal)
}

func IsJSONUnmarshalError(err error) bool {
	return errors.Is(err, ErrJSONUnmarshal)
}

func IsServiceUnreachableError(err error) bool {
	return errors.Is(err, ErrServiceUnreachable)
}

func IsServiceDiscoveryError(err error) bool {
	return errors.Is(err, ErrServiceDiscovery)
}

func IsSQLInjectionError(err error) bool {
	return errors.Is(err, ErrSQLInjection)
}

func IsRequestForgeryError(err error) bool {
	return errors.Is(err, ErrRequestForgery)
}

func IsSecretsLeakedError(err error) bool {
	return errors.Is(err, ErrSecretsLeaked)
}

func IsPIIBreachError(err error) bool {
	return errors.Is(err, ErrPIIBreach)
}

func IsTracerExporterError(err error) bool {
	return errors.Is(err, ErrTracerExporter)
}

func IsLogPipelineError(err error) bool {
	return errors.Is(err, ErrLogPipeline)
}

func IsMetricsCardinalityError(err error) bool {
	return errors.Is(err, ErrMetricsCardinality)
}

func IsBadRolloutError(err error) bool {
	return errors.Is(err, ErrBadRollout)
}

func IsSidecarCrashError(err error) bool {
	return errors.Is(err, ErrSidecarCrash)
}

func IsContainerImageError(err error) bool {
	return errors.Is(err, ErrContainerImage)
}
