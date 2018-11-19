package airvisual

import (
	"fmt"

	exception "github.com/blend/go-sdk/exception"
	request "github.com/blend/go-sdk/request"
)

// Client is an airvisual client
type Client struct {
	apiKey string
}

// New returns a new airvisual client
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// Location returns the data for a location
func (c *Client) Location(r *LocationRequest) (*Response, error) {
	err := r.Validate()
	if err != nil {
		return nil, err
	}
	req := request.Get(c.locationRequestURL(r))
	resp := &Response{}
	return resp, req.JSON(resp)
}

func (c *Client) locationRequestURL(r *LocationRequest) string {
	return fmt.Sprintf(LocationURLFormat, r.City, r.State, r.Country) + "&key=" + c.apiKey
}

// Validate validates the location request
func (r *LocationRequest) Validate() error {
	if r == nil {
		return exception.New("NilRequest")
	} else if len(r.City) == 0 {
		return exception.New("MissingCity")
	} else if len(r.State) == 0 {
		return exception.New("MissingState")
	} else if len(r.Country) == 0 {
		return exception.New("MissingCountry")
	}
	return nil
}
