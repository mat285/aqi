package cron

import "time"

// EverySecond returns a schedule that fires every second.
func EverySecond() Schedule {
	return IntervalSchedule{Every: 1 * time.Second}
}

// EveryMinute returns a schedule that fires every minute.
func EveryMinute() Schedule {
	return IntervalSchedule{Every: 1 * time.Minute}
}

// EveryHour returns a schedule that fire every hour.
func EveryHour() Schedule {
	return IntervalSchedule{Every: 1 * time.Hour}
}

// Every returns a schedule that fires every given interval.
func Every(interval time.Duration) Schedule {
	return IntervalSchedule{Every: interval}
}

// IntervalSchedule is as chedule that fires every given interval with an optional start delay.
type IntervalSchedule struct {
	Every      time.Duration
	StartDelay *time.Duration
}

// GetNextRunTime implements Schedule.
func (i IntervalSchedule) GetNextRunTime(after *time.Time) *time.Time {
	if after == nil {
		if i.StartDelay == nil {
			next := Now().Add(i.Every)
			return &next
		}
		next := Now().Add(*i.StartDelay).Add(i.Every)
		return &next
	}
	last := *after
	last = last.Add(i.Every)
	return &last
}
