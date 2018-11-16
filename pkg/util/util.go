package util

import (
	"fmt"
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
