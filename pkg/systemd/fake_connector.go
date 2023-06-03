package systemd

import (
	"time"
)

// FakeSystemdConnector is for testing the controller.
type FakeSystemdConnector struct {
	// Timestamp holds the method execution's timestamp.
	Timestamp map[string]time.Time
	// ServiceStarted checks whether the (fake) service is started.
	ServiceStarted map[string]bool
}

// CheckServiceStatus implements systemd.Connector
func (c *FakeSystemdConnector) CheckServiceStatus(serviceName string) error {
	c.Timestamp["CheckServiceStatus"] = time.Now()
	return nil
}

// StartService implements systemd.Connector
func (c *FakeSystemdConnector) StartService(serviceName string, preHook func() error, postHook func() error) error {
	c.ServiceStarted[serviceName] = true
	c.Timestamp["StartService"] = time.Now()
	return nil
}

// StopService implements systemd.Connector
func (c *FakeSystemdConnector) StopService(serviceName string) error {
	c.ServiceStarted[serviceName] = false
	c.Timestamp["StopService"] = time.Now()
	return nil
}

func NewFakeSystemdConnector() Connector {
	return &FakeSystemdConnector{
		Timestamp:      make(map[string]time.Time),
		ServiceStarted: make(map[string]bool),
	}
}
