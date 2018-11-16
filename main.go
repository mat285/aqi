package main

import (
	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
	"github.com/mat285/aqi/pkg/airvisual"
	config "github.com/mat285/aqi/pkg/config"
	"github.com/mat285/aqi/pkg/slack"
	"github.com/mat285/aqi/pkg/util"
)

func main() {
	agent := logger.All()
	conf, err := config.NewFromEnv()
	if err != nil {
		logger.FatalExit(err)
	}
	client := airvisual.New(conf.AirVisualAPIKey)
	req := &airvisual.LocationRequest{
		City:    "San%20Francisco",
		State:   "California",
		Country: "USA",
	}
	agent.SyncInfof("Sending request for San Francisco air data")
	resp, err := client.Location(req)
	if err != nil {
		agent.SyncFatalExit(err)
	}
	if resp.Status == airvisual.StatusFailed {
		agent.SyncFatalExit(exception.New("RequestFailed").WithMessagef("%v", resp))
	}
	aqi := resp.Data.Current.Pollution.AQI
	agent.SyncInfof("AQI: `%d`", aqi)

	channel := conf.GetSlackChannel("slack-bot-test")
	agent.SyncInfof("Notifying slack channel `%s`", channel)
	message := &slack.Message{
		Username:  util.SlackUsername,
		Text:      util.SlackMessageText(aqi),
		IconEmoji: util.EmojiForAQI(aqi),
		Channel:   channel,
	}
	err = slack.Notify(conf.SlackWebhook, message)
	if err != nil {
		logger.FatalExit(err)
	}
}
