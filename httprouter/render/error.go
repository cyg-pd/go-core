package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var fallbackErrorRender ErrorRenderFunc = func(err error) (Renderable, bool) {
	return internalServerError(err), true
}

func internalServerError(err error) Renderable {
	return Error{
		statusCode:       http.StatusInternalServerError,
		ErrorCode:        "internal_server_error",
		ErrorDescription: err.Error(),
	}
}

func SetFallbackErrorRender(h ErrorRenderFunc) { fallbackErrorRender = h }

type Error struct {
	statusCode int

	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *Error) SetStatusCode(c int) { e.statusCode = c }
func (e Error) StatusCode() int      { return e.statusCode }
func (e Error) Error() string {
	if e.ErrorDescription != "" {
		return e.ErrorDescription
	}
	return e.ErrorCode
}

func (e Error) Render(ctx *gin.Context) {
	Negotiate(ctx, e.statusCode, e)
}

func NewError(code, msg string, statusCode int) Error {
	return Error{
		statusCode:       statusCode,
		ErrorCode:        code,
		ErrorDescription: msg,
	}
}
