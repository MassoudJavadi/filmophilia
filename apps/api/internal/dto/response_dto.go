package dto

import "time"

type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
}

type MessageData struct {
	Message string `json:"message" example:"operation completed successfully"`
}

type CountData struct {
	Count int32 `json:"count" example:"12"`
}

type UnreadCountData struct {
	UnreadCount int32 `json:"unread_count" example:"3"`
}

type FollowStatusData struct {
	IsFollowing bool `json:"is_following" example:"true"`
}

type WatchlistStatusData struct {
	InWatchlist bool `json:"in_watchlist" example:"true"`
}

type ReactionMutationData struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code" example:"BAD_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

type BadRequestErrorResponse struct {
	Success bool            `json:"success" example:"false"`
	Error   BadRequestError `json:"error"`
}

type BadRequestError struct {
	Code    string `json:"code" example:"BAD_REQUEST"`
	Message string `json:"message" example:"invalid request"`
}

type UnauthorizedErrorResponse struct {
	Success bool                  `json:"success" example:"false"`
	Error   UnauthorizedErrorInfo `json:"error"`
}

type UnauthorizedErrorInfo struct {
	Code    string `json:"code" example:"UNAUTHORIZED"`
	Message string `json:"message" example:"authentication required"`
}

type ForbiddenErrorResponse struct {
	Success bool               `json:"success" example:"false"`
	Error   ForbiddenErrorInfo `json:"error"`
}

type ForbiddenErrorInfo struct {
	Code    string `json:"code" example:"FORBIDDEN"`
	Message string `json:"message" example:"you are not allowed to perform this action"`
}

type NotFoundErrorResponse struct {
	Success bool              `json:"success" example:"false"`
	Error   NotFoundErrorInfo `json:"error"`
}

type NotFoundErrorInfo struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"resource not found"`
}

type ConflictErrorResponse struct {
	Success bool              `json:"success" example:"false"`
	Error   ConflictErrorInfo `json:"error"`
}

type ConflictErrorInfo struct {
	Code    string `json:"code" example:"CONFLICT"`
	Message string `json:"message" example:"resource already exists"`
}

type InternalServerErrorResponse struct {
	Success bool                    `json:"success" example:"false"`
	Error   InternalServerErrorInfo `json:"error"`
}

type InternalServerErrorInfo struct {
	Code    string `json:"code" example:"INTERNAL_SERVER_ERROR"`
	Message string `json:"message" example:"internal server error"`
}

type TooManyRequestsErrorResponse struct {
	Success bool                     `json:"success" example:"false"`
	Error   TooManyRequestsErrorInfo `json:"error"`
}

type TooManyRequestsErrorInfo struct {
	Code       string `json:"code" example:"RATE_LIMIT_EXCEEDED"`
	Message    string `json:"message" example:"rate limit exceeded"`
	RetryAfter int    `json:"retry_after" example:"60"`
}
