package render

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func init() {
	RegisterErrorRender(isValidation)
}

type validationError struct {
	Error

	ErrorFields map[string]string `json:"fields"`
}

func (e validationError) Render(ctx *gin.Context) {
	Negotiate(ctx, e.Error.StatusCode(), e)
}

func isValidation(err error) (Renderable, bool) {
	var validate *validator.ValidationErrors
	if errors.As(err, &validate) {
		fields := make(map[string]string, len(*validate))

		for _, v := range *validate {
			fields[v.Field()] = v.ActualTag()
		}

		return validationError{
			Error:       NewError("bad_request", err.Error(), http.StatusBadRequest),
			ErrorFields: fields,
		}, true
	}
	return nil, false
}
