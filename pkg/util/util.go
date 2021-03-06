package util

import (
	"fmt"
	"io/ioutil"
	"strings"

	exception "github.com/blend/go-sdk/exception"
	logger "github.com/blend/go-sdk/logger"
	util "github.com/blendlabs/go-util"
	"github.com/mat285/aqi/pkg/airvisual"
	"github.com/mat285/aqi/pkg/config"
	"github.com/mat285/slack/slack"
)

const (
	// SlackUsername is the slack username
	SlackUsername = "AQI Bot"
	// SlackEmoji is the slack emoji
	SlackEmoji = ":cloud:"
	// HealthyEmoji is the healthy emoji
	HealthyEmoji = ":slightly_smiling_face:"
	// UnhealthyEmoji is the unhealthy emoji
	UnhealthyEmoji = ":mask:"
	// ToxicEmoji is the toxic emoji
	ToxicEmoji = ":skull_and_crossbones:"

	// CigarettesPerAQI is the number of cigarettes per point of aqi
	CigarettesPerAQI = 0.04631

	// CountryCodeUSA is the country code for USA
	CountryCodeUSA = "USA"
	// StateCodeCalifornia is the state code for california
	StateCodeCalifornia = "California"
)

var (
	// BlockedUsers are the users who need to ask nicely
	BlockedUsers = map[string]bool{}
)

// BlockUsersFromFile reads the newline delimited file to block the user ids
func BlockUsersFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return exception.New(err)
	}
	BlockUsers(strings.Split(string(data), " ")...)
	return nil
}

// BlockUsers blocks the given users
func BlockUsers(users ...string) {
	for _, u := range users {
		BlockedUsers[u] = true
	}
}

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
func SlackMessageText(aqi int, city string) string {
	return fmt.Sprintf("%s current AQI: `%d` %s", city, aqi, EmojiForAQI(aqi))
}

// LocationRequestFromText returns the location request from the text
func LocationRequestFromText(text string) *airvisual.LocationRequest {
	text = strings.TrimSpace(strings.ToLower(text))
	if strings.HasPrefix(text, "city ") {
		return CityAirVisualRequest(text)
	} else if strings.Contains(text, "sf") || strings.Contains(text, "san francisco") {
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

// CityAirVisualRequest returns the request for a city
func CityAirVisualRequest(text string) *airvisual.LocationRequest {
	text = strings.TrimSpace(strings.Trim(text, "city"))
	parts := SplitOnSpacePreserveQuotes(text)
	logger.All().Debugf("Parsed Input: %v", parts)
	if len(parts) < 3 {
		return nil
	}
	return &airvisual.LocationRequest{
		City:    util.String.ToTitleCase(parts[0]),
		State:   util.String.ToTitleCase(parts[1]),
		Country: util.String.ToTitleCase(parts[2]),
	}
}

// SanFranciscoAirVisualRequest returns the request for sf
func SanFranciscoAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "San Francisco",
		State:   StateCodeCalifornia,
		Country: CountryCodeUSA,
	}
}

// NewYorkAirVisualRequest returns the request for nyc
func NewYorkAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "New York",
		State:   "New York",
		Country: CountryCodeUSA,
	}
}

// LosAngelesAirVisualRequest returns the request for la
func LosAngelesAirVisualRequest() *airvisual.LocationRequest {
	return &airvisual.LocationRequest{
		City:    "Los Angeles",
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
	m := AQISlackMessage(-1, "")
	m.Text = "no"
	return m
}

// CigarettesSlackMessage returns the message for cigarettes
func CigarettesSlackMessage(aqi int, city string) *slack.Message {
	m := AQISlackMessage(aqi, city)
	m.Text = fmt.Sprintf("%s number of cigarettes: `%03f`", city, NumCigarettes(aqi))
	return m
}

// AQISlackMessage returns the message to send back for the aqi to slack
func AQISlackMessage(aqi int, city string) *slack.Message {
	return &slack.Message{
		Username:     SlackUsername,
		Text:         SlackMessageText(aqi, city),
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
	if resp.Status != airvisual.StatusSuccess {
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
	message := AQISlackMessage(aqi, req.City)
	message.Channel = channel
	return aqi, slack.Notify(c.SlackWebhook, message)
}
