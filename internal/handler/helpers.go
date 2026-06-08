package handler

import (
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// getUserID safely extracts user_id from gin context
func getUserID(c *gin.Context) (int64, bool) {
	val, ok := c.Get("user_id")
	if !ok {
		response.Error(c, response.ErrUnauthorized)
		return 0, false
	}
	id, ok := val.(int64)
	if !ok {
		response.Error(c, response.ErrUnauthorized)
		return 0, false
	}
	return id, true
}
