package cron

// Status is a status object
type Status struct {
	Jobs    []JobMeta
	Running map[string]JobInvocation
}
