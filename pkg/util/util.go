package util

import (
	"fmt"

	logger "github.com/blend/go-sdk/logger"
	exception "github.com/blendlabs/go-exception"
	"github.com/mat285/aqi/pkg/airvisual"
	"github.com/mat285/aqi/pkg/slack"

	"github.com/mat285/aqi/pkg/config"
)

const (
	SlackUsername  = "AQI Bot"
	HealthyEmoji   = ":slightly_smiling_face:"
	UnhealthyEmoji = ":mask"
	ToxicEmoji     = ":skull_and_crossbones:"
)

// EmojiForAQI returns the appropriate emohi for the aqi
func EmojiForAQI(aqi int) string {
	if aqi <= 50 {
		return HealthyEmoji
	} else if aqi <= 200 {
		return UnhealthyEmoji
	} else {
		return ToxicEmoji
	}
}

// SlackMessageText returns the text for a slack message of the aqi
func SlackMessageText(aqi int) string {
	return fmt.Sprintf("Current AQI: `%d` %s", aqi, EmojiForAQI(aqi))
}

// SanFranciscoAirVisualRequest returns the request for sf
func SanFranciscoAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "San%20Francisco",
		State:   "California",
		Country: "USA",
	}
}

// AQISlackMessage returns the message to send back for the aqi to slack
func AQISlackMessage(aqi int) *slack.Message {
	return &slack.Message{
		Username:     SlackUsername,
		Text:         SlackMessageText(aqi),
		IconEmoji:    EmojiForAQI(aqi),
		ResponseType: slack.ResponseTypeInChannel,
	}
}

// FetchAQI fetches the aqi from airvisual
func FetchAQI(c *config.Config, req *airvisual.LocationRequest, log *logger.Logger) (int, error) {
	client := airvisual.New(c.AirVisualAPIKey)
	log.SyncInfof("Sending request for air data")
	resp, err := client.Location(req)
	if err != nil {
		return -1, err
	}
	if resp.Status == airvisual.StatusFailed {
		return -1, exception.New("RequestFailed").WithMessagef("%v", resp)
	}
	return resp.Data.Current.Pollution.AQI, nil
}

// FetchAndSendAQIForConfig fetches aqi and sends it for the config
func FetchAndSendAQIForConfig(c *config.Config, req *airvisual.LocationRequest, log *logger.Logger) (int, error) {
	aqi, err := FetchAQI(c, req, log)
	if err != nil {
		return -1, err
	}
	log.SyncInfof("AQI: `%d`", aqi)

	channel := c.GetSlackChannel("slack-bot-test")
	log.SyncInfof("Notifying slack channel `%s`", channel)
	message := AQISlackMessage(aqi)
	message.Channel = channel
	return aqi, slack.Notify(c.SlackWebhook, message)
}
