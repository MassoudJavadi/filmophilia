package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
	DefaultPage  = 1
)

// Params contains validated pagination parameters
type Params struct {
	Limit  int32
	Offset int32
	Page   int32
}

// Parse extracts and validates pagination parameters from query string.
// Enforces MaxLimit to prevent abuse.
func Parse(c *gin.Context) Params {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))

	// Enforce limits
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	if page <= 0 {
		page = DefaultPage
	}

	offset := (page - 1) * limit

	return Params{
		Limit:  int32(limit),
		Offset: int32(offset),
		Page:   int32(page),
	}
}
