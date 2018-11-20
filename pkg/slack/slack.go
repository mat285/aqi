package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	exception "github.com/blend/go-sdk/exception"
	request "github.com/blend/go-sdk/request"
)

const (
	// ResponseTypeInChannel is the in channel response type
	ResponseTypeInChannel = "in_channel"

	// ParamTextKey is the slack text key
	ParamTextKey = "text"
	// ParamUserIDKey is the slack user key
	ParamUserIDKey = "user_id"

	// TimestampHeaderParam is the header for timestamp
	TimestampHeaderParam = "X-Slack-Request-Timestamp"
	// SignatureHeaderParam is the header for signature
	SignatureHeaderParam = "X-Slack-Signature"

	// EnvVarSignatureSecret is the secret for the signature
	EnvVarSignatureSecret = "SLACK_SIGNATURE_SECRET"
)

// Message is a message sent to slack.
type Message struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text"`
	Username     string `json:"username,omitempty"`
	UnfurlLinks  bool   `json:"unfurl_links"`
	IconEmoji    string `json:"icon_emoji,omitempty"`

	Channel string `json:"channel,omitempty"`
}

// Notify sends a slack hook.
func Notify(hook string, message *Message) error {
	hookURL, err := url.Parse(hook)
	if err != nil {
		return exception.New(err)
	}
	res, meta, err := request.New().AsPost().WithURL(hookURL).WithPostBodyAsJSON(message).StringWithMeta()
	if err != nil {
		return err
	}
	if meta.StatusCode > http.StatusOK {
		return exception.New(res)
	}
	return nil
}

// VerifyRequest verifies the request came from slack
func VerifyRequest(timestamp, body, digest, secret string) error {
	parts := strings.Split(digest, "=")
	if len(parts) != 2 {
		return exception.New("InvalidDigestError")
	}
	version := parts[0]
	sigBase := fmt.Sprintf("%s:%s:%s", version, timestamp, body)
	hasher := hmac.New(sha256.New, []byte(secret))
	_, err := hasher.Write([]byte(sigBase))
	if err != nil {
		return exception.New(err)
	}
	mac := hasher.Sum(nil)
	expected, err := hex.DecodeString(parts[1])
	if err != nil {
		return exception.New(err)
	}
	if !hmac.Equal(mac, expected) {
		return exception.New("SignatureInvalid")
	}
	return nil
}
