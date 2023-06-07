package mariadb

import (
	"fmt"
	"time"
)

type FakeMariaDBConnector struct {
	Timestamp        map[string]time.Time
	ReadOnlyVariable bool
	MasterConfig     MasterInstance
}

func NewFakeMariaDBConnector() Connector {
	return &FakeMariaDBConnector{
		Timestamp:        make(map[string]time.Time),
		ReadOnlyVariable: false,
		MasterConfig:     MasterInstance{},
	}
}

// ChangeMasterTo implements mariadb.Connector
func (c *FakeMariaDBConnector) ChangeMasterTo(master MasterInstance) error {
	c.Timestamp["ChangeMasterTo"] = time.Now()
	c.MasterConfig = master
	return nil
}

// CheckBoolVariableIsON implements mariadb.Connector
func (c *FakeMariaDBConnector) CheckBoolVariableIsON(variableName string) bool {
	c.Timestamp["CheckBoolVariableIsON"] = time.Now()
	return c.ReadOnlyVariable
}

// ResetAllReplicas implements mariadb.Connector
func (c *FakeMariaDBConnector) ResetAllReplicas() error {
	c.Timestamp["ResetAllRelicas"] = time.Now()
	return nil
}

// ShowReplicationStatus implements mariadb.Connector
func (c *FakeMariaDBConnector) ShowReplicationStatus() (ReplicationStatus, error) {
	c.Timestamp["ShowReplicationStatus"] = time.Now()
	status := ReplicationStatus{
		ReplicationStatusSlaveIORunning:  "Yes",
		ReplicationStatusSlaveSQLRunning: "Yes",
	}
	return status, nil
}

// StartReplica implements mariadb.Connector
func (c *FakeMariaDBConnector) StartReplica() error {
	c.Timestamp["StartReplica"] = time.Now()
	return nil
}

// StopReplica implements mariadb.Connector
func (c *FakeMariaDBConnector) StopReplica() error {
	c.Timestamp["StopReplica"] = time.Now()
	return nil
}

// TurnOffBoolVariable implements mariadb.Connector
func (c *FakeMariaDBConnector) TurnOffBoolVariable(variableName string) error {
	c.Timestamp[fmt.Sprintf("TurnOffBoolVariable(%s)", variableName)] = time.Now()
	if variableName == ReadOnlyVariableName {
		c.ReadOnlyVariable = false
	}

	return nil
}

// TurnOnBoolVariable implements mariadb.Connector
func (c *FakeMariaDBConnector) TurnOnBoolVariable(variableName string) error {
	c.Timestamp[fmt.Sprintf("TurnOnBoolVariable(%s)", variableName)] = time.Now()
	if variableName == ReadOnlyVariableName {
		c.ReadOnlyVariable = true
	}
	return nil
}

// CreateDatabase implements mariadb.Connector
func (c *FakeMariaDBConnector) CreateDatabase(dbName string) error {
	c.Timestamp[fmt.Sprintf("CreateDatabase(%s)", dbName)] = time.Now()
	return nil
}

// CreateIDTable implements mariadb.Connector
func (c *FakeMariaDBConnector) CreateIDTable(dbName string, tableName string) error {
	c.Timestamp[fmt.Sprintf("CreateIDTable(%s, %s)", dbName, tableName)] = time.Now()
	return nil
}

// DeleteRecords implements mariadb.Connector
func (c *FakeMariaDBConnector) DeleteRecords(dbName string, tableName string) error {
	c.Timestamp[fmt.Sprintf("DeleteRecords(%s, %s)", dbName, tableName)] = time.Now()
	return nil
}

// InsertIDRecord implements mariadb.Connector
func (c *FakeMariaDBConnector) InsertIDRecord(dbName string, tableName string, id int) error {
	c.Timestamp[fmt.Sprintf("InsertIDRecord(%s, %s, %d)", dbName, tableName, id)] = time.Now()
	return nil
}

// FakeMariaDBFailWriteTestDataConnector is the mariadb connector that fails to write testdata.
type FakeMariaDBFailWriteTestDataConnector struct {
}

func NewFakeMariaDBFailWriteTestDataConnector() Connector {
	return &FakeMariaDBFailWriteTestDataConnector{}
}

// ChangeMasterTo implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) ChangeMasterTo(master MasterInstance) error {
	return nil
}

// CheckBoolVariableIsON implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) CheckBoolVariableIsON(variableName string) bool {
	return true
}

// ResetAllReplicas implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) ResetAllReplicas() error {
	return nil
}

// ShowReplicationStatus implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) ShowReplicationStatus() (ReplicationStatus, error) {
	status := ReplicationStatus{
		ReplicationStatusSlaveIORunning:  "Yes",
		ReplicationStatusSlaveSQLRunning: "No", // data-inconsistency
	}

	return status, nil
}

// StartReplica implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) StartReplica() error {
	return nil
}

// StopReplica implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) StopReplica() error {
	return nil
}

// TurnOffBoolVariable implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) TurnOffBoolVariable(variableName string) error {
	return nil
}

// TurnOnBoolVariable implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) TurnOnBoolVariable(variableName string) error {
	return nil
}

// CreateDatabase implements mariadb.Connector
func (*FakeMariaDBFailWriteTestDataConnector) CreateDatabase(dbName string) error {
	return fmt.Errorf("failed to create %s database", dbName)
}

// CreateIDTable implements mariadb.Connector
func (*FakeMariaDBFailWriteTestDataConnector) CreateIDTable(dbName string, tableName string) error {
	return nil
}

// DeleteRecords implements mariadb.Connector
func (*FakeMariaDBFailWriteTestDataConnector) DeleteRecords(dbName string, tableName string) error {
	return nil
}

// InsertIDRecord implements mariadb.Connector
func (*FakeMariaDBFailWriteTestDataConnector) InsertIDRecord(dbName string, tableName string, id int) error {
	return nil
}
