package server

import (
	"github.com/blend/go-sdk/web"
)

// Config is the config for a slack server
type Config struct {
	web.Config           `json:",inline"`
	SlackSignatureSecret string `json:"slackSignatureSecret"`
	AcknowledgeOnVerify  bool   `json:"acknowledgeOnVerify"`
}
