package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/blend/go-sdk/env"
	exception "github.com/blend/go-sdk/exception"
)

// Config configures the project
type Config struct {
	AirVisualAPIKey string `yaml:"airvisualAPIKey" env:"AIRVISUAL_API_KEY"`
	SlackWebhook    string `yaml:"slackWebhook" env:"SLACK_WEBHOOK"`
	SlackChannel    string `yaml:"slackChannel" env:"SLACK_CHANNEL"`
}

// NewFromFile returns a new config from a file
func NewFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, exception.New(err)
	}
	c := &Config{}
	return c, exception.New(json.Unmarshal(data, c))
}

// NewFromEnv returns a new config from the environment
func NewFromEnv() (*Config, error) {
	c := &Config{}
	return c, env.Env().ReadInto(c)
}

// Validate validates the config
func (c *Config) Validate() error {
	if c == nil {
		return exception.New("NilConfig")
	} else if len(c.AirVisualAPIKey) == 0 {
		return exception.New("MissingAPIKey")
	} else if len(c.SlackWebhook) == 0 {
		return exception.New("MissingSlackWebhook")
	}
	return nil
}

// GetSlackChannel returns the slack channel
func (c *Config) GetSlackChannel(defaults ...string) string {
	if len(c.SlackChannel) > 0 {
		return c.SlackChannel
	}
	for _, d := range defaults {
		if len(d) > 0 {
			return d
		}
	}
	return ""
}
