package web

import "net/http"

var (
	// NoContent is a static result.
	NoContent NoContentResult
)

// NoContentResult returns a no content response.
type NoContentResult struct{}

// Render renders a static result.
func (ncr NoContentResult) Render(ctx *Ctx) error {
	ctx.Response().WriteHeader(http.StatusNoContent)
	return nil
}
