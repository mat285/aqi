package web

import (
	"net/http"
)

var (
	// JSON is a static singleton json result provider.
	JSON JSONResultProvider
	// assert it implements result provider.
	_ ResultProvider = (*JSONResultProvider)(nil)
)

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct{}

// NotFound returns a service response.
func (jrp JSONResultProvider) NotFound() Result {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (jrp JSONResultProvider) NotAuthorized() Result {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (jrp JSONResultProvider) InternalError(err error) Result {
	if err != nil {
		return resultWithLoggedError(&JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response:   err.Error(),
		}, err)
	}
	return resultWithLoggedError(&JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response:   "Internal Server Error",
	}, err)
}

// BadRequest returns a service response.
func (jrp JSONResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &JSONResult{
			StatusCode: http.StatusBadRequest,
			Response:   err.Error(),
		}
	}
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response:   "Bad Request",
	}
}

// OK returns a service response.
func (jrp JSONResultProvider) OK() Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Status returns a plaintext result.
func (jrp JSONResultProvider) Status(statusCode int, response ...interface{}) Result {
	return &JSONResult{
		StatusCode: statusCode,
		Response:   ResultOrDefault(http.StatusText(statusCode), response...),
	}
}

// Result returns a json response.
func (jrp JSONResultProvider) Result(response interface{}) Result {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   response,
	}
}
