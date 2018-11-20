package util

import (
	"fmt"
	"strings"

	exception "github.com/blend/go-sdk/exception"
	logger "github.com/blend/go-sdk/logger"
	"github.com/mat285/aqi/pkg/airvisual"
	"github.com/mat285/aqi/pkg/slack"

	"github.com/mat285/aqi/pkg/config"
)

const (
	SlackUsername  = "AQI Bot"
	SlackEmoji     = ":cloud:"
	HealthyEmoji   = ":slightly_smiling_face:"
	UnhealthyEmoji = ":mask:"
	ToxicEmoji     = ":skull_and_crossbones:"

	CigarettesPerAQI = 0.04631

	CountryCodeUSA      = "USA"
	StateCodeCalifornia = "California"
)

var (
	BlockedUsers = map[string]bool{}
)

// IsBlocked returns if the user is blocked
func IsBlocked(user string) bool {
	blocked, ok := BlockedUsers[user]
	return ok && blocked
}

// NumCigarettes returns the number of cigarettes for the aqi
func NumCigarettes(aqi int) float32 {
	return float32(aqi) * CigarettesPerAQI
}

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

// LocationRequestFromText returns the location request from the text
func LocationRequestFromText(text string) *airvisual.LocationRequest {
	text = strings.ToLower(text)
	if strings.Contains(text, "sf") || strings.Contains(text, "san francisco") {
		return SanFranciscoAirVisualRequest()
	} else if strings.Contains(text, "nyc") || strings.Contains(text, "new york") {
		return NewYorkAirVisualRequest()
	} else if strings.Contains(text, "seattle") {
		return SeattleAirVisualRequest()
	} else if strings.Contains(text, " la ") || strings.TrimSpace(text) == "la" || strings.Contains(text, "los angeles") {
		return LosAngelesAirVisualRequest()
	}
	return SanFranciscoAirVisualRequest()
}

// SanFranciscoAirVisualRequest returns the request for sf
func SanFranciscoAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "San%20Francisco",
		State:   StateCodeCalifornia,
		Country: CountryCodeUSA,
	}
}

// NewYorkAirVisualRequest returns the request for nyc
func NewYorkAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "New%20York",
		State:   "New%20York",
		Country: CountryCodeUSA,
	}
}

// LosAngelesAirVisualRequest returns the request for la
func LosAngelesAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "Los%20Angeles",
		State:   StateCodeCalifornia,
		Country: CountryCodeUSA,
	}
}

// SeattleAirVisualRequest returns the request for seattle
func SeattleAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "Seattle",
		State:   "Washington",
		Country: CountryCodeUSA,
	}
}

// BlockedSlackMessage returns the message to a blocked user
func BlockedSlackMessage() *slack.Message {
	m := AQISlackMessage(-1)
	m.Text = "no"
	return m
}

// CigarettesSlackMessage returns the message for cigarettes
func CigarettesSlackMessage(aqi int) *slack.Message {
	m := AQISlackMessage(aqi)
	m.Text = fmt.Sprintf("Number of cigarettes: `%03f`", NumCigarettes(aqi))
	return m
}

// AQISlackMessage returns the message to send back for the aqi to slack
func AQISlackMessage(aqi int) *slack.Message {
	return &slack.Message{
		Username:     SlackUsername,
		Text:         SlackMessageText(aqi),
		IconEmoji:    SlackEmoji,
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
