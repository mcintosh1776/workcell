package workcell

// IsTerminalJobState returns true if the job state is terminal (succeeded or failed).
func IsTerminalJobState(state JobState) bool {
	switch state {
	case JobSucceeded, JobFailed:
		return true
	default:
		return false
	}
}