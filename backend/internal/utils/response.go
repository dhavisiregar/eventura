package utils

import "github.com/gin-gonic/gin"

// Success sends a standard { data, meta? } envelope.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func SuccessWithMeta(c *gin.Context, status int, data any, meta any) {
	c.JSON(status, gin.H{"data": data, "meta": meta})
}

// Error sends a standard { error: { message } } envelope.
func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"message": message}})
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewPagination(page, limit int, total int64) Pagination {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}
	return Pagination{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}
