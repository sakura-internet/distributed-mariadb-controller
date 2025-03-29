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

type FakeMariaDBFailedReplicationConnector struct {
}

func NewFakeMariaDBFailedReplicationConnector() Connector {
	return &FakeMariaDBFailedReplicationConnector{}
}

// ChangeMasterTo implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) ChangeMasterTo(master MasterInstance) error {
	return nil
}

// IsReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) IsReadOnly() bool {
	return true
}

// ResetAllReplicas implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) ResetAllReplicas() error {
	return nil
}

// ShowReplicationStatus implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) ShowReplicationStatus() (ReplicationStatus, error) {
	status := ReplicationStatus{
		ReplicationStatusSlaveIORunning:  "Yes",
		ReplicationStatusSlaveSQLRunning: "No", // data-inconsistency
	}

	return status, nil
}

// StartReplica implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) StartReplica() error {
	return nil
}

// StopReplica implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) StopReplica() error {
	return nil
}

// TurnOffReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) TurnOffReadOnly() error {
	return nil
}

// TurnOnReadOnly implements mariadb.Connector
func (c *FakeMariaDBFailedReplicationConnector) TurnOnReadOnly() error {
	return nil
}

// CreateDatabase implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) CreateDatabase(dbName string) error {
	return nil
}

// CreateIDTable implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) CreateIDTable(dbName string, tableName string) error {
	return nil
}

// DeleteRecords implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) DeleteRecords(dbName string, tableName string) error {
	return nil
}

// InsertIDRecord implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) InsertIDRecord(dbName string, tableName string, id int) error {
	return nil
}

// RemoveMasterInfo implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) RemoveMasterInfo() error {
	return nil
}

// RemoveRelayInfo implements mariadb.Connector
func (*FakeMariaDBFailedReplicationConnector) RemoveRelayInfo() error {
	return nil
}
