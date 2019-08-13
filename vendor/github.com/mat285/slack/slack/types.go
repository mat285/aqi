package slack

// Message is a message sent to slack.
type Message struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text"`
	Username     string `json:"username,omitempty"`
	UnfurlLinks  bool   `json:"unfurl_links"`
	IconEmoji    string `json:"icon_emoji,omitempty"`

	Channel string `json:"channel,omitempty"`
}

// SlashCommandRequest is the request struct sent from slack on a slash command
type SlashCommandRequest struct {
	Token          string `json:"token"`
	TeamID         string `json:"team_id"`
	TeamDomain     string `json:"team_domain"`
	EnterpriseID   string `json:"enterprise_id"`
	EnterpriseName string `json:"enterprise_name"`
	ChannelID      string `json:"channel_id"`
	ChannelName    string `json:"channel_name"`
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	Command        string `json:"command"`
	Text           string `json:"text"`
	ResponseURL    string `json:"response_url"`
	TriggerID      string `json:"trigger_id"`
}
