package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

func SetRequestContextWithTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		newRequest := c.Request.WithContext(ctx)
		c.Request = newRequest

		c.Next()
	}
}
