package cron

import (
	"time"
)

// Schedule is a type that provides a next runtime after a given previous runtime.
type Schedule interface {
	// GetNextRuntime should return the next runtime after a given previous runtime. If `after` is <nil> it should be assumed
	// the job hasn't run yet. If <nil> is returned by the schedule it is inferred that the job should not run again.
	GetNextRunTime(*time.Time) *time.Time
}
