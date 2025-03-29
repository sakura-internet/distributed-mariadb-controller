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
	"os"
	"strings"
	"time"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/command"
	"golang.org/x/exp/slog"
)

const (
	readOnlyVariableName = "read_only"
)

var (
	mysqlCommandTimeout = 5 * time.Second
)

// Connector is an interface that communicates with mariadb.
type Connector interface {
	// about readonly variable mechanism
	IsReadOnly() bool
	TurnOnReadOnly() error
	TurnOffReadOnly() error

	// about replication
	ChangeMasterTo(master MasterInstance) error
	StartReplica() error
	StopReplica() error
	ResetAllReplicas() error
	ShowReplicationStatus() (ReplicationStatus, error)

	// about operation for DB health check
	CreateDatabase(dbName string) error
	CreateIDTable(dbName string, tableName string) error
	InsertIDRecord(dbName string, tableName string, id int) error
	DeleteRecords(dbName string, tableName string) error

	// remove master info or relay info
	RemoveMasterInfo() error
	RemoveRelayInfo() error
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &mySQLCommandConnector{logger: logger}
}

type mySQLCommandConnector struct {
	logger *slog.Logger
}

// CreateIDTable implements Connector
func (c *mySQLCommandConnector) CreateIDTable(dbName string, tableName string) error {
	createCmd := fmt.Sprintf("create table if not exists %s.%s(id int)", dbName, tableName)
	if _, err := c.runMysqlCommand(createCmd); err != nil {
		return fmt.Errorf("failed to create %s table on %s table: %w", dbName, tableName, err)
	}

	return nil
}

// CreateDatabase implements Connector
func (c *mySQLCommandConnector) CreateDatabase(dbName string) error {
	createCmd := fmt.Sprintf("create database if not exists %s", dbName)
	if _, err := c.runMysqlCommand(createCmd); err != nil {
		return fmt.Errorf("failed to create %s database: %w", dbName, err)
	}

	return nil
}

// DeleteRecords implements Connector
func (c *mySQLCommandConnector) DeleteRecords(dbName string, tableName string) error {
	deleteCmd := fmt.Sprintf("delete from %s.%s", dbName, tableName)
	if _, err := c.runMysqlCommand(deleteCmd); err != nil {
		return fmt.Errorf("failed to delete records from %s.%s: %w", dbName, tableName, err)
	}

	return nil
}

// InsertIDRecord implements Connector
func (c *mySQLCommandConnector) InsertIDRecord(dbName string, tableName string, id int) error {
	insertCmd := fmt.Sprintf("insert into %s.%s values(%d)", dbName, tableName, id)
	if _, err := c.runMysqlCommand(insertCmd); err != nil {
		return fmt.Errorf("failed to insert id record to %s.%s: %w", dbName, tableName, err)
	}

	return nil
}

// IsReadOnly implements Connector
func (c *mySQLCommandConnector) IsReadOnly() bool {
	name := "mysql"
	args := []string{"-s", "-N", "-e", fmt.Sprintf("show variables like \"%s\"", readOnlyVariableName)}
	c.logger.Debug("execute command", "name", name, "args", args, "callerFn", "CheckBoolVariableIsON")

	out, err := command.RunWithTimeout(mysqlCommandTimeout, name, args...)
	if err != nil {
		c.logger.Debug("failed to show variable", "name", readOnlyVariableName, "error", err)
		return false
	}

	s := string(out)
	return strings.Contains(s, readOnlyVariableName) && strings.Contains(s, "ON")
}

// TurnOffReadOnly implements Connector
func (c *mySQLCommandConnector) TurnOffReadOnly() error {
	setCmd := fmt.Sprintf("set global %s=0", readOnlyVariableName)
	if _, err := c.runMysqlCommand(setCmd); err != nil {
		return fmt.Errorf("failed to set '%s' variable to 0: %w", readOnlyVariableName, err)
	}

	return nil
}

// TurnOnReadOnly implements Connector
func (c *mySQLCommandConnector) TurnOnReadOnly() error {
	setCmd := fmt.Sprintf("set global %s=1", readOnlyVariableName)
	if _, err := c.runMysqlCommand(setCmd); err != nil {
		return fmt.Errorf("failed to set '%s' variable to 1: %w", readOnlyVariableName, err)
	}

	return nil
}

// StopReplica implements Connector
func (c *mySQLCommandConnector) StartReplica() error {
	if _, err := c.runMysqlCommand("start replica"); err != nil {
		return fmt.Errorf("failed to start replica: %w", err)
	}

	return nil
}

// StopReplica implements Connector
func (c *mySQLCommandConnector) StopReplica() error {
	if _, err := c.runMysqlCommand("stop replica"); err != nil {
		return fmt.Errorf("failed to stop replica: %w", err)
	}

	return nil
}

// ResetAllReplicas implements Connector
func (c *mySQLCommandConnector) ResetAllReplicas() error {
	if _, err := c.runMysqlCommand("reset replica all"); err != nil {
		return fmt.Errorf("failed to reset all replicas: %w", err)
	}

	return nil
}

// ChangeMasterTo implements Connector
func (c *mySQLCommandConnector) ChangeMasterTo(
	master MasterInstance,
) error {
	changeMasterOpts := []string{}
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_host = \"%s\"", master.Host))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_port = %d", master.Port))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_user = \"%s\"", master.User))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_password = \"%s\"", master.Password))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_use_gtid = %s", master.UseGTID))

	cmd := fmt.Sprintf("change master to %s", strings.Join(changeMasterOpts, ", "))
	if out, err := c.runMysqlCommand(cmd); err != nil {
		c.logger.Debug("changeMasterTo", "output", string(out))
		return fmt.Errorf("failed to change master to: %w", err)
	}

	return nil
}

// ShowReplicationStatus implements Connector
func (c *mySQLCommandConnector) ShowReplicationStatus() (ReplicationStatus, error) {
	out, err := c.runMysqlCommand("show replica status\\G")
	if err != nil {
		return ReplicationStatus{}, fmt.Errorf("failed to show replica status: %w", err)
	}

	return parseShowReplicaStatusOutput(string(out)), nil
}

// runMysqlCommand executes specified mysql command with timeout and logging
func (c *mySQLCommandConnector) runMysqlCommand(mysqlcmd string) ([]byte, error) {
	name := "mysql"
	args := []string{"-e", mysqlcmd}

	c.logger.Debug("execute command", "name", name, "args", args)
	return command.RunWithTimeout(mysqlCommandTimeout, name, args...)
}

func (c *mySQLCommandConnector) RemoveMasterInfo() error {
	_, err := os.Stat(MasterInfoFilePath)

	// do nothing if file is not found
	if err != nil {
		return nil
	}

	// delete if file exists
	return os.Remove(MasterInfoFilePath)
}

func (c *mySQLCommandConnector) RemoveRelayInfo() error {
	_, err := os.Stat(RelayInfoFilePath)

	// do nothing if file is not found
	if err != nil {
		return nil
	}

	// delete if file exists
	return os.Remove(RelayInfoFilePath)
}

// parseShowReplicaStatusOutput parses the output of the "mysql -e 'show replica status \G'".
func parseShowReplicaStatusOutput(out string) ReplicationStatus {
	m := ReplicationStatus{}

	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, ":") {
			continue
		}

		kv := strings.Split(line, ":")
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		m[key] = value
	}

	return m
}
