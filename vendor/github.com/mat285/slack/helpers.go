package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	exception "github.com/blendlabs/go-exception"
	request "github.com/blendlabs/go-request"
)

// VerifyRequest verifies the request came from slack
func VerifyRequest(timestamp, body, digest string, secret []byte) error {
	parts := strings.Split(digest, "=")
	if len(parts) != 2 {
		return exception.New(ErrInvalidDigest)
	}
	version := parts[0]
	sigBase := fmt.Sprintf("%s:%s:%s", version, timestamp, body)
	hasher := hmac.New(sha256.New, secret)
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
		return exception.New(ErrSignatureInvalid)
	}
	return nil
}

// Notify sends a slack hook.
func Notify(hook string, message *Message) error {
	res, meta, err := request.New().AsPost().WithURL(hook).WithPostBodyAsJSON(message).StringWithMeta()
	if err != nil {
		return err
	}
	if meta.StatusCode > http.StatusOK {
		return exception.New(res)
	}
	return nil
}

// UnmarshalSlashCommandBody unmarshals the form encoded data into the struct
func UnmarshalSlashCommandBody(body []byte) (*SlashCommandRequest, error) {
	strbody := strings.Replace(string(body), "\n&", "&", -1)
	vals, err := url.ParseQuery(strbody)
	if err != nil {
		return nil, exception.New(err)
	}
	data, err := json.Marshal(vals)
	if err != nil {
		return nil, exception.New(err)
	}
	req := SlashCommandRequest{}
	if err != nil {
		return nil, exception.New(err)
	}
	return &req, exception.New(json.Unmarshal(data, &req))
}