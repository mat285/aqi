package slack

import (
	"net/http"
	"net/url"

	exception "github.com/blend/go-sdk/exception"
	request "github.com/blend/go-sdk/request"
)

const (
	ResponseTypeInChannel = "in_channel"

	ParamTextKey   = "text"
	ParamUserIDKey = "user_id"
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
