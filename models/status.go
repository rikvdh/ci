package models

const (
	// StatusUnknown indicates a unknown build status
	StatusUnknown = "unknown"
	// StatusNew indicates a new unbuilt build
	StatusNew = "new"
	// StatusBusy indicates a currently in progress build
	StatusBusy = "busy"
	// StatusFailed indicates a failed build status
	StatusFailed = "failed"
	// StatusPassed indicates a succesful build
	StatusPassed = "passed"
	// StatusError indicates an error occurred during build
	StatusError = "error"
)
