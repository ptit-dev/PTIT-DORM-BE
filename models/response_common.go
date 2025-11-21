package models

import "net/http"

// Response represents a standard API response
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"total_page"`
}


// SuccessResponse creates a successful response with data
func SuccessResponse(data any) Response {
	return Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    data,
	}
}

// SuccessResponseWithMessage creates a successful response with custom message
func SuccessResponseWithMessage(message string, data any) Response {
	return Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}



// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Time    string `json:"time"`
}
