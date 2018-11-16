package cron

// NOTE: ALL TIMES ARE IN UTC. JUST USE UTC.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/blend/go-sdk/async"
	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/logger"
)

// New returns a new job manager.
func New() *JobManager {
	jm := JobManager{
		latch:   &async.Latch{},
		jobs:    map[string]*JobMeta{},
		running: map[string]*JobInvocation{},
	}
	jm.schedulerWorker = async.NewInterval(jm.runDueJobs, DefaultHeartbeatInterval)
	jm.killHangingTasksWorker = async.NewInterval(jm.killHangingJobs, DefaultHeartbeatInterval)
	return &jm
}

// NewFromConfig returns a new job manager from a given config.
func NewFromConfig(cfg *Config) *JobManager {
	return New()
}

// NewFromEnv returns a new job manager from the environment.
func NewFromEnv() (*JobManager, error) {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewFromConfig(cfg), nil
}

// MustNewFromEnv returns a new job manager from the environment.
func MustNewFromEnv() *JobManager {
	cfg, err := NewConfigFromEnv()
	if err != nil {
		panic(err)
	}
	return NewFromConfig(cfg)
}

// JobManager is the main orchestration and job management object.
type JobManager struct {
	sync.Mutex

	latch *async.Latch

	tracer Tracer
	log    *logger.Logger

	schedulerWorker        *async.Interval
	killHangingTasksWorker *async.Interval

	jobs    map[string]*JobMeta
	running map[string]*JobInvocation
}

// WithLogger sets the logger and returns a reference to the job manager.
func (jm *JobManager) WithLogger(log *logger.Logger) *JobManager {
	jm.log = log
	return jm
}

// Logger returns the diagnostics agent.
func (jm *JobManager) Logger() *logger.Logger {
	return jm.log
}

// WithTracer sets the manager's tracer.
func (jm *JobManager) WithTracer(tracer Tracer) *JobManager {
	jm.tracer = tracer
	return jm
}

// Tracer returns the manager's tracer.
func (jm *JobManager) Tracer() Tracer {
	return jm.tracer
}

// Latch returns the internal latch.
func (jm *JobManager) Latch() *async.Latch {
	return jm.latch
}

// --------------------------------------------------------------------------------
// Core Methods
// --------------------------------------------------------------------------------

// LoadJobs loads a variadic list of jobs.
func (jm *JobManager) LoadJobs(jobs ...Job) error {
	jm.Lock()
	defer jm.Unlock()

	var err error
	for _, job := range jobs {
		err = jm.loadJobUnsafe(job)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadJob loads a job.
func (jm *JobManager) LoadJob(job Job) error {
	jm.Lock()
	defer jm.Unlock()
	return jm.loadJobUnsafe(job)
}

// DisableJobs disables a variadic list of job names.
func (jm *JobManager) DisableJobs(jobNames ...string) error {
	jm.Lock()
	defer jm.Unlock()

	var err error
	for _, jobName := range jobNames {
		err = jm.setJobDisabledUnsafe(jobName, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// DisableJob stops a job from running but does not unload it.
func (jm *JobManager) DisableJob(jobName string) error {
	jm.Lock()
	defer jm.Unlock()

	return jm.setJobDisabledUnsafe(jobName, true)
}

// EnableJobs enables a variadic list of job names.
func (jm *JobManager) EnableJobs(jobNames ...string) error {
	var err error
	for _, jobName := range jobNames {
		err = jm.setJobDisabledUnsafe(jobName, false)
		if err != nil {
			return err
		}
	}
	return nil
}

// EnableJob enables a job that has been disabled.
func (jm *JobManager) EnableJob(jobName string) error {
	jm.Lock()
	defer jm.Unlock()

	return jm.setJobDisabledUnsafe(jobName, false)
}

// HasJob returns if a jobName is loaded or not.
func (jm *JobManager) HasJob(jobName string) (hasJob bool) {
	jm.Lock()
	defer jm.Unlock()
	_, hasJob = jm.jobs[jobName]
	return
}

// Job returns a job metadata by name.
func (jm *JobManager) Job(jobName string) (job *JobMeta, err error) {
	jm.Lock()
	defer jm.Unlock()
	if jobMeta, hasJob := jm.jobs[jobName]; hasJob {
		job = jobMeta
	} else {
		err = exception.New(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
	}
	return
}

// IsJobDisabled returns if a job is disabled.
func (jm *JobManager) IsJobDisabled(jobName string) (value bool) {
	jm.Lock()
	defer jm.Unlock()

	if job, hasJob := jm.jobs[jobName]; hasJob {
		value = job.Disabled
		if job.EnabledProvider != nil {
			value = value || !job.EnabledProvider()
		}
	}
	return
}

// IsJobRunning returns if a task is currently running.
func (jm *JobManager) IsJobRunning(jobName string) (isRunning bool) {
	jm.Lock()
	defer jm.Unlock()
	_, isRunning = jm.running[jobName]
	return
}

// RunJobs runs a variadic list of job names.
func (jm *JobManager) RunJobs(jobNames ...string) error {
	jm.Lock()
	defer jm.Unlock()
	for _, jobName := range jobNames {
		if jobMeta, ok := jm.jobs[jobName]; ok {
			jm.runJobUnsafe(jobMeta)
		} else {
			return exception.New(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
		}
	}
	return nil
}

// RunJob runs a job by jobName on demand.
func (jm *JobManager) RunJob(jobName string) error {
	jm.Lock()
	defer jm.Unlock()

	if job, ok := jm.jobs[jobName]; ok {
		jm.runJobUnsafe(job)
		return nil
	}
	return exception.New(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
}

// RunAllJobs runs every job that has been loaded in the JobManager at once.
func (jm *JobManager) RunAllJobs() {
	jm.Lock()
	defer jm.Unlock()
	for _, jobMeta := range jm.jobs {
		jm.runJobUnsafe(jobMeta)
	}
}

// CancelJob cancels (sends the cancellation signal) to a running job.
func (jm *JobManager) CancelJob(jobName string) (err error) {
	jm.Lock()
	defer jm.Unlock()

	job, ok := jm.running[jobName]
	if !ok {
		err = exception.New(ErrJobNotFound).WithMessagef("job: %s", jobName)
		return
	}

	job.Elapsed = Since(job.StartTime)
	job.Err = exception.New(ErrJobCancelled)
	job.Cancel()
	return
}

// Status returns a status object.
func (jm *JobManager) Status() *Status {
	jm.Lock()
	defer jm.Unlock()
	status := Status{
		Running: map[string]JobInvocation{},
	}
	for _, meta := range jm.jobs {
		status.Jobs = append(status.Jobs, *meta)
	}
	for name, job := range jm.running {
		status.Running[name] = *job
	}
	return &status
}

//
// Life Cycle
//

// Start begins the schedule runner for a JobManager.
func (jm *JobManager) Start() error {
	if !jm.latch.CanStart() {
		return fmt.Errorf("already started")
	}
	jm.latch.Starting()
	jm.schedulerWorker.Start()
	jm.killHangingTasksWorker.Start()
	jm.latch.Started()
	return nil
}

// Stop stops the schedule runner for a JobManager.
func (jm *JobManager) Stop() error {
	if !jm.latch.CanStop() {
		return fmt.Errorf("already stopped")
	}
	jm.latch.Stopping()
	jm.schedulerWorker.Stop()
	jm.killHangingTasksWorker.Stop()
	jm.latch.Stopped()
	return nil
}

// NotifyStarted returns the started notification channel.
func (jm *JobManager) NotifyStarted() <-chan struct{} {
	return jm.latch.NotifyStarted()
}

// NotifyStopped returns the stopped notification channel.
func (jm *JobManager) NotifyStopped() <-chan struct{} {
	return jm.latch.NotifyStopped()
}

// IsRunning returns if the job manager is running.
// It serves as an authoritative healthcheck.
func (jm *JobManager) IsRunning() bool {
	return jm.latch.IsRunning() && jm.schedulerWorker.IsRunning() && jm.killHangingTasksWorker.IsRunning()
}

// --------------------------------------------------------------------------------
// lifecycle methods
// --------------------------------------------------------------------------------

func (jm *JobManager) runDueJobs() error {
	jm.Lock()
	defer jm.Unlock()
	now := Now()
	for _, jobMeta := range jm.jobs {
		if !jobMeta.NextRunTime.IsZero() && jobMeta.NextRunTime.Before(now) {
			jm.runJobUnsafe(jobMeta)
		}
	}
	return nil
}

func (jm *JobManager) killHangingJobs() (err error) {
	jm.Lock()
	defer jm.Unlock()

	var effectiveTimeout time.Time
	var now time.Time
	var t1, t2 time.Time

	for jobName, ji := range jm.running {
		if ji.Timeout.IsZero() {
			return
		}
		now = Now()
		if jobMeta, ok := jm.jobs[jobName]; ok {
			nextRuntime := jobMeta.NextRunTime
			t1 = ji.Timeout
			t2 = nextRuntime
			effectiveTimeout = Min(t1, t2)
		} else {
			effectiveTimeout = ji.Timeout
		}
		if effectiveTimeout.Before(now) {
			jm.killHangingJob(ji)
		}
	}
	return nil
}

func (jm *JobManager) killHangingJob(ji *JobInvocation) {
	ji.Cancel()
}

//
// these assume a lock is held
//

func (jm *JobManager) runJobUnsafe(jobMeta *JobMeta) {
	if !jm.jobCanRun(jobMeta) {
		return
	}

	now := Now()
	jobMeta.NextRunTime = jm.scheduleNextRuntime(jobMeta.Schedule, Optional(now))

	start := Now()
	ctx, cancel := jm.createContextWithCancel()

	ji := JobInvocation{
		ID:        NewJobInvocationID(),
		Name:      jobMeta.Name,
		StartTime: start,
		JobMeta:   jobMeta,
		Context:   ctx,
		Cancel:    cancel,
	}

	if jobMeta.TimeoutProvider != nil {
		if timeout := jobMeta.TimeoutProvider(); timeout > 0 {
			ji.Timeout = start.Add(timeout)
		}
	}

	jm.running[ji.Name] = &ji
	go jm.execute(WithJobInvocation(ctx, &ji), &ji)
}

func (jm *JobManager) execute(ctx context.Context, ji *JobInvocation) {
	var err error
	var tf TraceFinisher
	defer func() {
		if tf != nil {
			tf.Finish(ctx)
		}

		jm.Lock()
		if _, ok := jm.running[ji.Name]; ok {
			delete(jm.running, ji.Name)
		}
		jm.Unlock()

		ji.Elapsed = Since(ji.StartTime)
		ji.Err = err

		if err != nil && IsJobCancelled(err) {
			jm.onCancelled(ctx, ji)
		} else if ji.Err != nil {
			jm.onFailure(ctx, ji)
		} else {
			jm.onComplete(ctx, ji)
		}
		ji.JobMeta.Last = ji
	}()
	if jm.tracer != nil {
		ctx, tf = jm.tracer.Start(ctx)
	}

	jm.onStart(ctx, ji)

	select {
	case <-ctx.Done():
		err = ErrJobCancelled
	case err = <-jm.safeAsyncExec(ctx, ji.JobMeta.Job):
		return
	}
}

func (jm *JobManager) safeAsyncExec(ctx context.Context, job Job) chan error {
	errors := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errors <- exception.New(r)
			}
		}()
		errors <- job.Execute(ctx)
	}()
	return errors
}

// --------------------------------------------------------------------------------
// Utility Methods
// --------------------------------------------------------------------------------

// LoadJob adds a job to the manager.
func (jm *JobManager) loadJobUnsafe(j Job) error {
	jobName := j.Name()

	if _, hasJob := jm.jobs[jobName]; hasJob {
		return exception.New(ErrJobAlreadyLoaded).WithMessagef("job: %s", j.Name())
	}

	meta := &JobMeta{
		Name: jobName,
		Job:  j,
	}

	if typed, ok := j.(ScheduleProvider); ok {
		meta.Schedule = typed.Schedule()
		meta.NextRunTime = jm.scheduleNextRuntime(meta.Schedule, nil)
	}

	if typed, ok := j.(TimeoutProvider); ok {
		meta.TimeoutProvider = typed.Timeout
	} else {
		meta.TimeoutProvider = func() time.Duration { return 0 }
	}

	if typed, ok := j.(EnabledProvider); ok {
		meta.EnabledProvider = typed.Enabled
	} else {
		meta.EnabledProvider = func() bool { return DefaultEnabled }
	}

	if typed, ok := j.(SerialProvider); ok {
		meta.SerialProvider = typed.Serial
	} else {
		meta.SerialProvider = func() bool { return DefaultSerial }
	}

	if typed, ok := j.(ShouldTriggerListenersProvider); ok {
		meta.ShouldTriggerListenersProvider = typed.ShouldTriggerListeners
	} else {
		meta.ShouldTriggerListenersProvider = func() bool { return DefaultShouldTriggerListeners }
	}

	if typed, ok := j.(ShouldWriteOutputProvider); ok {
		meta.ShouldWriteOutputProvider = typed.ShouldWriteOutput
	} else {
		meta.ShouldWriteOutputProvider = func() bool { return DefaultShouldWriteOutput }
	}

	jm.jobs[jobName] = meta
	return nil
}

func (jm *JobManager) scheduleNextRuntime(schedule Schedule, after *time.Time) time.Time {
	if schedule != nil {
		return Deref(schedule.GetNextRunTime(after))
	}
	return time.Time{}
}

func (jm *JobManager) setJobDisabledUnsafe(jobName string, disabled bool) error {
	if job, hasJob := jm.jobs[jobName]; hasJob {
		job.Disabled = disabled
		return nil
	}
	return exception.New(ErrJobNotLoaded).WithMessagef("job: %s", jobName)
}

func (jm *JobManager) createContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// jobCanRun returns whether a job can be executed
func (jm *JobManager) jobCanRun(job *JobMeta) bool {
	if job.Disabled {
		return false
	}
	if job.EnabledProvider != nil {
		if !job.EnabledProvider() {
			return false
		}
	}

	if job.SerialProvider != nil && job.SerialProvider() {
		_, hasTask := jm.running[job.Name]
		if hasTask {
			return false
		}
	}
	return true
}

func (jm *JobManager) onStart(ctx context.Context, ji *JobInvocation) {
	if jm.log != nil && ji.JobMeta.ShouldTriggerListenersProvider() {
		event := NewEvent(FlagStarted, ji.Name).WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider())
		jm.log.SubContext(ji.ID).Trigger(event)
	}
	if typed, ok := ji.JobMeta.Job.(OnStartReceiver); ok {
		typed.OnStart(ctx)
	}
}

func (jm *JobManager) onCancelled(ctx context.Context, ji *JobInvocation) {
	if jm.log != nil && ji.JobMeta.ShouldTriggerListenersProvider() {
		event := NewEvent(FlagCancelled, ji.Name).
			WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider()).
			WithElapsed(ji.Elapsed)
		jm.log.SubContext(ji.ID).Trigger(event)
	}
	if typed, ok := ji.JobMeta.Job.(OnCancellationReceiver); ok {
		typed.OnCancellation(ctx)
	}
}

func (jm *JobManager) onComplete(ctx context.Context, ji *JobInvocation) {
	if jm.log != nil && ji.JobMeta.ShouldTriggerListenersProvider() {
		event := NewEvent(FlagComplete, ji.Name).
			WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider()).
			WithElapsed(ji.Elapsed)
		jm.log.SubContext(ji.ID).Trigger(event)
	}
	if typed, ok := ji.JobMeta.Job.(OnCompleteReceiver); ok {
		typed.OnComplete(ctx)
	}

	if ji.JobMeta.Last != nil && ji.JobMeta.Last.Err != nil {
		if jm.log != nil {
			event := NewEvent(FlagFixed, ji.Name).
				WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider()).
				WithElapsed(ji.Elapsed)
			jm.log.SubContext(ji.ID).Trigger(event)
		}

		if typed, ok := ji.JobMeta.Job.(OnFixedReceiver); ok {
			typed.OnFixed(ctx)
		}
	}
}

func (jm *JobManager) onFailure(ctx context.Context, ji *JobInvocation) {
	if jm.log != nil && ji.JobMeta.ShouldTriggerListenersProvider() {
		event := NewEvent(FlagFailed, ji.Name).
			WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider()).
			WithElapsed(ji.Elapsed).
			WithErr(ji.Err)

		jm.log.SubContext(ji.ID).Trigger(event)
	}
	if ji.Err != nil {
		jm.err(ji.Err)
	}
	if typed, ok := ji.JobMeta.Job.(OnFailureReceiver); ok {
		typed.OnFailure(ctx)
	}
	if ji.JobMeta.Last != nil && ji.JobMeta.Last.Err == nil {
		if jm.log != nil {
			event := NewEvent(FlagBroken, ji.Name).
				WithIsWritable(ji.JobMeta.ShouldWriteOutputProvider()).
				WithElapsed(ji.Elapsed)
			jm.log.SubContext(ji.ID).Trigger(event)
		}

		if typed, ok := ji.JobMeta.Job.(OnBrokenReceiver); ok {
			typed.OnBroken(ctx)
		}
	}
}

//
// logging helpers
//

func (jm *JobManager) err(err error) {
	if err != nil && jm.log != nil {
		jm.log.Error(err)
	}
}

func (jm *JobManager) fatal(err error) {
	if err != nil && jm.log != nil {
		jm.log.Fatal(err)
	}
}

func (jm *JobManager) errorf(format string, args ...interface{}) {
	if jm.log != nil {
		jm.log.SyncErrorf(format, args...)
	}
}

func (jm *JobManager) debugf(format string, args ...interface{}) {
	if jm.log != nil {
		jm.log.SyncDebugf(format, args...)
	}
}
