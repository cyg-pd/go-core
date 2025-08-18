package render

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/render"
)

var NegotiateOffered = []string{
	binding.MIMEJSON,
	binding.MIMEXML, binding.MIMEXML2,
	binding.MIMEPROTOBUF,
	binding.MIMEMSGPACK, binding.MIMEMSGPACK2,
	binding.MIMEYAML, binding.MIMEYAML2,
}

// Negotiate serializes the given data as Negotiate into the response body.
func Negotiate(c *gin.Context, code int, obj any) {
	switch c.NegotiateFormat(NegotiateOffered...) {
	case binding.MIMEJSON:
		c.Render(code, render.JSON{Data: obj})

	case binding.MIMEXML, binding.MIMEXML2:
		c.Render(code, render.XML{Data: obj})

	case binding.MIMEPROTOBUF:
		c.Render(code, render.ProtoBuf{Data: obj})

	case binding.MIMEMSGPACK, binding.MIMEMSGPACK2:
		c.Render(code, render.MsgPack{Data: obj})

	case binding.MIMEYAML, binding.MIMEYAML2:
		c.Render(code, render.YAML{Data: obj})

	default:
		_ = c.AbortWithError(http.StatusNotAcceptable, errors.New("the accepted formats are not offered by the server"))
	}
}
