package cron

import (
	"sync"
	"time"
)

// Immediately Returns a schedule that casues a job to run immediately on start,
// with an optional subsequent schedule.
func Immediately() *ImmediateSchedule {
	return &ImmediateSchedule{}
}

// ImmediateSchedule fires immediately with an optional continuation schedule.
type ImmediateSchedule struct {
	sync.Mutex

	didRun bool
	then   Schedule
}

// Then allows you to specify a subsequent schedule after the first run.
func (i *ImmediateSchedule) Then(then Schedule) Schedule {
	i.then = then
	return i
}

// GetNextRunTime implements Schedule.
func (i *ImmediateSchedule) GetNextRunTime(after *time.Time) *time.Time {
	i.Lock()
	defer i.Unlock()

	if !i.didRun {
		i.didRun = true
		return Optional(Now())
	}
	if i.then != nil {
		return i.then.GetNextRunTime(after)
	}
	return nil
}
