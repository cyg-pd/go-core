package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var fallbackErrorRender ErrorRenderFunc = func(err error) (Renderable, bool) {
	return &defaultErrorRender{err: err}, true
}

func SetFallbackErrorRender(h ErrorRenderFunc) { fallbackErrorRender = h }

type defaultErrorRender struct{ err error }

func (d *defaultErrorRender) Render(ctx *gin.Context) {
	Negotiate(ctx, http.StatusInternalServerError, d.err.Error())
}

var _ Renderable = (*defaultErrorRender)(nil)
