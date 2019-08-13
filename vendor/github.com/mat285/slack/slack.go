package slack

import (
	"io/ioutil"
	"net/http"
)

// Slack manages slack work
type Slack struct {
	RequestSigningSecret []byte
}

// New returns a new slack manager
func New(secret []byte) *Slack {
	return &Slack{
		RequestSigningSecret: secret,
	}
}

// VerifyRequest verifies the request and returns the posted data
func (s *Slack) VerifyRequest(r *http.Request) (*SlashCommandRequest, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	timestamp := r.Header.Get(TimestampHeaderParam)
	sig := r.Header.Get(SignatureHeaderParam)
	err = VerifyRequest(timestamp, string(body), sig, s.RequestSigningSecret)
	if err != nil {
		return nil, err
	}
	return UnmarshalSlashCommandBody(body)
}
