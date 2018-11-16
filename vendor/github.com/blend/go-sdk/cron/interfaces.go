package cron

import (
	"context"
	"time"
)

/*
A note on the naming conventions for the below interfaces.

MethodName[Receiver|Provider] is the general pattern.

"Receiver" indicates the function will be called by the manager.

"Proivder" indicates the function will be called and is expected to return a specific value.

They're mostly the same except
*/

// ScheduleProvider returns a schedule for the job.
type ScheduleProvider interface {
	Schedule() Schedule
}

// TimeoutProvider is an interface that allows a task to be timed out.
type TimeoutProvider interface {
	Timeout() time.Duration
}

// StatusProvider is an interface that allows a task to report its status.
type StatusProvider interface {
	Status() string
}

// SerialProvider is an optional interface that prohibits
// a task from running if another instance of the task is currently running.
type SerialProvider interface {
	Serial() bool
}

// ShouldTriggerListenersProvider is a type that enables or disables logger listeners.
type ShouldTriggerListenersProvider interface {
	ShouldTriggerListeners() bool
}

// ShouldWriteOutputProvider is a type that enables or disables logger output for events.
type ShouldWriteOutputProvider interface {
	ShouldWriteOutput() bool
}

// EnabledProvider is an optional interface that will allow jobs to control if they're enabled.
type EnabledProvider interface {
	Enabled() bool
}

// OnStartReceiver is an interface that allows a task to be signaled when it has started.
type OnStartReceiver interface {
	OnStart(context.Context)
}

// OnCancellationReceiver is an interface that allows a task to be signaled when it has been canceled.
type OnCancellationReceiver interface {
	OnCancellation(context.Context)
}

// OnCompleteReceiver is an interface that allows a task to be signaled when it has been completed.
type OnCompleteReceiver interface {
	OnComplete(context.Context)
}

// OnFailureReceiver is an interface that allows a task to be signaled when it has been completed.
type OnFailureReceiver interface {
	OnFailure(context.Context)
}

// OnBrokenReceiver is an interface that allows a job to be signaled when it is a failure that followed
// a previous success.
type OnBrokenReceiver interface {
	OnBroken(context.Context)
}

// OnFixedReceiver is an interface that allows a jbo to be signaled when is a success that followed
// a previous failure.
type OnFixedReceiver interface {
	OnFixed(context.Context)
}
