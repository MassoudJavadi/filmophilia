package response

import (
	"net/http"

	"github.com/MassoudJavadi/filmophilia/api/internal/dto"
	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data interface{}) {
	Success(c, http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	Success(c, http.StatusCreated, data)
}

func Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, dto.SuccessResponse{
		Success: true,
		Data:    data,
	})
}

func Message(c *gin.Context, status int, message string) {
	Success(c, status, dto.MessageData{Message: message})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, dto.ErrorResponse{
		Success: false,
		Error: dto.ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

func AbortError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, dto.ErrorResponse{
		Success: false,
		Error: dto.ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

func AbortRateLimitError(c *gin.Context, retryAfter int) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.TooManyRequestsErrorResponse{
		Success: false,
		Error: dto.TooManyRequestsErrorInfo{
			Code:       "RATE_LIMIT_EXCEEDED",
			Message:    "rate limit exceeded",
			RetryAfter: retryAfter,
		},
	})
}
