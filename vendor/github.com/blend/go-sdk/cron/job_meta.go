package cron

import (
	"time"
)

// JobMeta is runtime metadata for a job.
type JobMeta struct {
	Name        string         `json:"name"`
	Job         Job            `json:"-"`
	Disabled    bool           `json:"disabled"`
	NextRunTime time.Time      `json:"nextRunTime"`
	Last        *JobInvocation `json:"last"`

	// these are used at runtime and not serialized with the status.
	Schedule                       Schedule             `json:"-"`
	EnabledProvider                func() bool          `json:"-"`
	SerialProvider                 func() bool          `json:"-"`
	TimeoutProvider                func() time.Duration `json:"-"`
	ShouldTriggerListenersProvider func() bool          `json:"-"`
	ShouldWriteOutputProvider      func() bool          `json:"-"`
}
