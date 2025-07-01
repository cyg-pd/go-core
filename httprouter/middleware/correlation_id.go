package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if t := trace.SpanContextFromContext(c.Request.Context()); t.IsValid() {
			c.Header(
				"X-Correlation-ID",
				t.TraceID().String(),
			)
		}
	}
}
