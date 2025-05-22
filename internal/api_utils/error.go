package api_utils

import (
	"github.com/go-chi/render"
	"net/http"
)

type ErrResponse struct {
	Err            error `json:"-"`          // low-level runtime error
	HTTPStatusCode int   `json:"statusCode"` // http response status code

	StatusText string `json:"status"`          // http status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     http.StatusText(http.StatusBadRequest),
		ErrorText:      err.Error(),
	}
}

func ErrInternalError(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     http.StatusText(http.StatusInternalServerError),
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{
	HTTPStatusCode: http.StatusNotFound,
	StatusText:     http.StatusText(http.StatusNotFound),
}

var ErrForbidden = &ErrResponse{
	HTTPStatusCode: http.StatusForbidden,
	StatusText:     http.StatusText(http.StatusForbidden),
}

var ErrUnauthorized = &ErrResponse{
	HTTPStatusCode: http.StatusUnauthorized,
	StatusText:     http.StatusText(http.StatusUnauthorized),
}
