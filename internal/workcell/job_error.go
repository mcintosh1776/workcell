package workcell

import "strings"

// HasBackendError reports whether a job has backend infrastructure error details.
func HasBackendError(job Job) bool {
  return strings.TrimSpace(job.Error) != ""
}