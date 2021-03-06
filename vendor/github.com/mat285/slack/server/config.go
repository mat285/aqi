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

// Status is the status of the server
type Status struct {
	Ready bool `json:"ready"`
}
