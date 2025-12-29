package api

// routeHandlers contains all the handlers for different route types
type routeHandlers struct {
	projectHandler  projectHandler
	blogPostHandler blogPostHandler
}

// ErrorResponse represents an error response from the API
// @Description Error response structure
type ErrorResponse struct {
	Error   string `json:"error" example:"Internal Server Error"`
	Status  string `json:"status" example:"error"`
	Field   string `json:"field,omitempty" example:"title"`
	Details string `json:"details,omitempty" example:"Additional error details"`
	Cause   string `json:"cause,omitempty" example:"Underlying error cause"`
}
