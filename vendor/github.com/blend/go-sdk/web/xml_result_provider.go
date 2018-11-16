package web

import (
	"net/http"
)

var (
	// XML is a static singleton xml result provider.
	XML XMLResultProvider

	// assert xml implements result provider.
	_ ResultProvider = XML
)

// XMLResultProvider are context results for api methods.
type XMLResultProvider struct{}

// NotFound returns a service response.
func (xrp XMLResultProvider) NotFound() Result {
	return &XMLResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (xrp XMLResultProvider) NotAuthorized() Result {
	return &XMLResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (xrp XMLResultProvider) InternalError(err error) Result {
	return resultWithLoggedError(&XMLResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err,
	}, err)
}

// BadRequest returns a service response.
func (xrp XMLResultProvider) BadRequest(err error) Result {
	if err != nil {
		return &XMLResult{
			StatusCode: http.StatusBadRequest,
			Response:   err,
		}
	}
	return &XMLResult{
		StatusCode: http.StatusBadRequest,
		Response:   "Bad Request",
	}
}

// OK returns a service response.
func (xrp XMLResultProvider) OK() Result {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Status returns a plaintext result.
func (xrp XMLResultProvider) Status(statusCode int, response ...interface{}) Result {
	return &XMLResult{
		StatusCode: statusCode,
		Response:   ResultOrDefault(http.StatusText(statusCode), response...),
	}
}

// Result returns an xml response.
func (xrp XMLResultProvider) Result(result interface{}) Result {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   result,
	}
}
