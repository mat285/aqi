package cron

import "time"

// EveryHourOnTheHour returns a schedule that fires every 60 minutes on the 00th minute.
func EveryHourOnTheHour() Schedule {
	return OnTheHourAt{}
}

// EveryHourAt returns a schedule that fires every hour at a given minute.
func EveryHourAt(minute, second int) Schedule {
	return OnTheHourAt{Minute: minute, Second: second}
}

// OnTheHourAt is a schedule that fires every hour on the given minute.
type OnTheHourAt struct {
	Minute int
	Second int
}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheHourAt) GetNextRunTime(after *time.Time) *time.Time {
	var returnValue time.Time
	now := Now()
	if after == nil {
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(now) {
			returnValue = returnValue.Add(time.Hour)
		}
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), o.Minute, o.Second, 0, time.UTC)
		if returnValue.Before(*after) {
			returnValue = returnValue.Add(time.Hour)
		}
	}
	return &returnValue
}
