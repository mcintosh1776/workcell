package workcell

// IsTerminalJobState returns true when state is a terminal Workcell job state.
func IsTerminalJobState(state JobState) bool {
	switch state {
	case JobSucceeded, JobFailed:
		return true
	case JobQueued, JobRunning:
		return false
	default:
		return false
	}
}