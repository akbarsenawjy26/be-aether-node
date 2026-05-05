package response

import (
	"github.com/labstack/echo/v4"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasMore    bool  `json:"has_more"`
}

type PaginatedResponse struct {
	Success    bool       `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func Success(c echo.Context, status int, data interface{}, message string) error {
	return c.JSON(status, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func SuccessWithPagination(c echo.Context, status int, data interface{}, pagination *Pagination) error {
	return c.JSON(status, PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
	})
}

func Error(c echo.Context, status int, code string, message string) error {
	return c.JSON(status, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}
