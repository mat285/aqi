package web

import "github.com/blend/go-sdk/util"

// HSTSConfig are hsts options.
type HSTSConfig struct {
	Enabled           *bool `json:"enabled" yaml:"enabled"`
	MaxAgeSeconds     int   `json:"maxAgeSeconds" yaml:"maxAgeSeconds"`
	IncludeSubDomains *bool `json:"includeSubDomains" yaml:"includeSubDomains"`
	Preload           *bool `json:"preload" yaml:"preload"`
}

// GetEnabled returns if hsts should be enabled.
func (h HSTSConfig) GetEnabled(defaults ...bool) bool {
	return util.Coalesce.Bool(h.Enabled, DefaultHSTS, defaults...)
}

// GetMaxAgeSeconds returns the max age seconds.
func (h HSTSConfig) GetMaxAgeSeconds(defaults ...int) int {
	return util.Coalesce.Int(h.MaxAgeSeconds, DefaultHSTSMaxAgeSeconds, defaults...)
}

// GetIncludeSubDomains returns if hsts should include sub-domains.
func (h HSTSConfig) GetIncludeSubDomains(defaults ...bool) bool {
	return util.Coalesce.Bool(h.IncludeSubDomains, DefaultHSTSIncludeSubDomains, defaults...)
}

// GetPreload returns if hsts should apply before requests.
func (h HSTSConfig) GetPreload(defaults ...bool) bool {
	return util.Coalesce.Bool(h.Preload, DefaultHSTSPreload, defaults...)
}
