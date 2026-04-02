package xerrors

import "errors"

var (
	// SchedulerErrEmpty is returned when the scheduler has no instances to return
	SchedulerErrEmpty = errors.New("scheduler is empty")
)
