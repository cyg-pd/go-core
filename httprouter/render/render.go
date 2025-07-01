package render

import (
	"github.com/gin-gonic/gin"
)

type Renderable interface {
	Render(ctx *gin.Context)
}

func RenderError(ctx *gin.Context, err error) {
	if e, ok := err.(Renderable); ok {
		e.Render(ctx)
		return
	}

	for _, v := range mapper {
		if r, ok := v(err); ok {
			r.Render(ctx)
			return
		}
	}

	if r, ok := fallbackErrorRender(err); ok {
		r.Render(ctx)
		return
	}

	internalServerError(err).Render(ctx)
}
