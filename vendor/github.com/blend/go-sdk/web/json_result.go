package web

import "github.com/blend/go-sdk/webutil"

// JSONResult is a json result.
type JSONResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (jr *JSONResult) Render(ctx *Ctx) error {
	return webutil.WriteJSON(ctx.Response(), jr.StatusCode, jr.Response)
}
