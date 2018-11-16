package cron

import "time"

// OnceAtUTC returns a schedule that fires once at a given time.
// It will never fire again unless reloaded.
func OnceAtUTC(t time.Time) Schedule {
	return OnceAtUTCSchedule{Time: t}
}

// OnceAtUTCSchedule is a schedule.
type OnceAtUTCSchedule struct {
	Time time.Time
}

// GetNextRunTime returns the next runtime.
func (oa OnceAtUTCSchedule) GetNextRunTime(after *time.Time) *time.Time {
	if after == nil {
		return &oa.Time
	}
	if oa.Time.After(*after) {
		return &oa.Time
	}
	return nil
}
