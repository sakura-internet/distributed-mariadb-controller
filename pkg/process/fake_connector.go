package process

import (
	"time"
)

// FakeProcessControlConnector is for testing the controller.
type FakeProcessControlConnector struct {
	// Timestamp holds the method execution's timestamp.
	Timestamp map[string]time.Time
	// ProcessLived checks whether the (fake) process is lived.
	ProcessLived map[string]bool
}

// KillProcessWithFullName implements process.ProcessControlConnector
func (c *FakeProcessControlConnector) KillProcessWithFullName(processName string) error {
	c.ProcessLived[processName] = false
	c.Timestamp["KillProcessWithFullName"] = time.Now()

	return nil
}

func NewFakeProcessControlConnector() ProcessControlConnector {
	return &FakeProcessControlConnector{
		Timestamp:    make(map[string]time.Time),
		ProcessLived: make(map[string]bool),
	}
}
