// Copyright 2025 The distributed-mariadb-controller Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// ResetAllReplicas implements mariadb.Connector
func (c *FakeMariaDBConnector) ResetAllReplicas() error {
	c.Timestamp["ResetAllReplicas"] = time.Now()
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

// IsReadOnly implements mariadb.Connector
func (c *FakeMariaDBConnector) IsReadOnly() bool {
	c.Timestamp["IsReadOnly"] = time.Now()
	return c.ReadOnlyVariable
}

// TurnOffReadOnly implements mariadb.Connector
func (c *FakeMariaDBConnector) TurnOffReadOnly() error {
	c.Timestamp["TurnOffReadOnly"] = time.Now()
	c.ReadOnlyVariable = false
	return nil
}

// TurnOnReadOnly implements mariadb.Connector
func (c *FakeMariaDBConnector) TurnOnReadOnly() error {
	c.Timestamp["TurnOnReadOnly"] = time.Now()
	c.ReadOnlyVariable = true
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

func (c *FakeMariaDBConnector) RemoveMasterInfo() error {
	return nil
}

func (c *FakeMariaDBConnector) RemoveRelayInfo() error {
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

// IsReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) IsReadOnly() bool {
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

// TurnOffReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) TurnOffReadOnly() error {
	return nil
}

// TurnOnReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailWriteTestDataConnector) TurnOnReadOnly() error {
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

func (c *FakeMariaDBFailWriteTestDataConnector) RemoveMasterInfo() error {
	return nil
}

func (c *FakeMariaDBFailWriteTestDataConnector) RemoveRelayInfo() error {
	return nil
}
