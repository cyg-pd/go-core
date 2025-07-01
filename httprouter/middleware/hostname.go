package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

func Hostname() gin.HandlerFunc {
	name, err := os.Hostname()
	if err != nil {
		return func(*gin.Context) {}
	}
	return func(c *gin.Context) {
		c.Header("x-host-name", name)
	}
}
