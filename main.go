package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/blend/go-sdk/cron"

	"github.com/blend/go-sdk/env"
	request "github.com/blend/go-sdk/request"
	exception "github.com/blendlabs/go-exception"
)

const (
	urlFormat = "http://api.airvisual.com/v2/city?city=San%20Francisco&state=California&country=USA&key="

	username = "AQI Bot"

	goodEmoji = ":slightly_smiling_face:"

	badEmoji = ":mask"

	veryBadEmoji = ":skull_and_crossbones:"
)

func main() {

	// manager := cron.New()
	// err := manager.LoadJob(&job{})
	// if err != nil {
	// 	panic(err)
	// }
	// err = manager.Start()
	// if err != nil {
	// 	panic(err)
	// }
	// select {}

	err := doJob()
	if err != nil {
		panic(err)
	}
}

type job struct{}

func (j *job) Name() string {
	return "get_aqi"
}

func (j *job) Schedule() cron.Schedule {
	return cron.EveryMinute()
}

func (j *job) Execute(ctx context.Context) error {
	return doJob()
}

func doJob() error {
	req := request.Get(urlFormat + env.Env().String("API_KEY"))
	resp := &airVisualResponse{}
	err := req.JSON(resp)
	if err != nil {
		return err
	}
	if resp.Status != "success" {
		return exception.New("Request Error")
	}

	message := &SlackMessage{
		Username:  username,
		IconEmoji: getEmoji(resp.Data.Current.Pollution.AQIUS),
		Channel:   env.Env().String("SLACK_CHANNEL", "slack-bot-test"),
		Text:      fmt.Sprintf("Current AQI: `%d`", resp.Data.Current.Pollution.AQIUS),
	}

	err = SlackNotify(env.Env().String("SLACK_WEBHOOK"), message)
	if err != nil {
		return err
	}
	return nil
}

func getEmoji(aqi int) string {
	if aqi <= 50 {
		return goodEmoji
	} else if aqi <= 200 {
		return badEmoji
	} else {
		return veryBadEmoji
	}
}

type airVisualResponse struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	Current current `json:"current"`
}

type current struct {
	Pollution pollution `json:"pollution"`
}

type pollution struct {
	AQIUS int `json:"aqius"`
}

// SlackMessage is a message sent to slack.
type SlackMessage struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text"`
	Username     string `json:"username,omitempty"`
	UnfurlLinks  bool   `json:"unfurl_links"`
	IconEmoji    string `json:"icon_emoji,omitempty"`

	Channel string `json:"channel,omitempty"`
}

// SlackNotify sends a slack hook.
func SlackNotify(hook string, message *SlackMessage) error {
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
