// Copyright 2023 The distributed-mariadb-controller Authors
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
	"strings"

	"github.com/sakura-internet/distributed-mariadb-controller/pkg/bash"
	"golang.org/x/exp/slog"
)

// Connector is an interface that communicates with mariadb.
type Connector interface {
	/*
	 * about variable mechanism
	 */

	// CheckBoolVariableIsON checks whether the <variableName> is ON.
	CheckBoolVariableIsON(variableName string) bool

	TurnOnBoolVariable(variableName string) error

	TurnOffBoolVariable(variableName string) error

	/*
	 * about replication
	 */
	ChangeMasterTo(master MasterInstance) error

	StartReplica() error

	StopReplica() error

	ResetAllReplicas() error

	ShowReplicationStatus() (ReplicationStatus, error)

	/*
	 *
	 */

	CreateDatabase(dbName string) error

	CreateIDTable(dbName string, tableName string) error

	InsertIDRecord(dbName string, tableName string, id int) error

	DeleteRecords(dbName string, tableName string) error
}

func NewDefaultConnector(logger *slog.Logger) Connector {
	return &MySQLCommandConnector{Logger: logger}
}

type MySQLCommandConnector struct {
	Logger *slog.Logger
}

// CreateIDTable implements Connector
func (c *MySQLCommandConnector) CreateIDTable(dbName string, tableName string) error {
	createCmd := fmt.Sprintf("create table if not exists %s.%s(id int)", dbName, tableName)
	cmd := fmt.Sprintf("timeout -s 9 5 mysql -e '%s'", createCmd)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "CreateIDTable")
	if _, err := bash.RunCommand(cmd); err != err {
		return fmt.Errorf("failed to create %s table on %s table: %w", dbName, tableName, err)
	}

	return nil
}

// CreateDatabase implements Connector
func (c *MySQLCommandConnector) CreateDatabase(dbName string) error {
	createCmd := fmt.Sprintf("create database if not exists %s", dbName)
	cmd := fmt.Sprintf("timeout -s 9 5 mysql -e '%s'", createCmd)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "CreateDatabase")
	if _, err := bash.RunCommand(cmd); err != err {
		return fmt.Errorf("failed to create %s database: %w", dbName, err)
	}

	return nil
}

// DeleteRecords implements Connector
func (c *MySQLCommandConnector) DeleteRecords(dbName string, tableName string) error {
	deleteCmd := fmt.Sprintf("delete from %s.%s", dbName, tableName)
	cmd := fmt.Sprintf("timeout -s 9 5 mysql -e '%s'", deleteCmd)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "DeleteRecords")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to delete records from %s.%s: %w", dbName, tableName, err)
	}

	return nil
}

// InsertIDRecord implements Connector
func (c *MySQLCommandConnector) InsertIDRecord(dbName string, tableName string, id int) error {
	insertCmd := fmt.Sprintf("insert into %s.%s values(%d)", dbName, tableName, id)
	cmd := fmt.Sprintf("timeout -s 9 5 mysql -e '%s'", insertCmd)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "InsertIDRecord")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to insert id record to %s.%s: %w", dbName, tableName, err)
	}

	return nil
}

// CheckBoolVariableIsON implements Connector
func (c *MySQLCommandConnector) CheckBoolVariableIsON(variableName string) bool {
	cmd := fmt.Sprintf("mysql -s -N -e 'show variables like \"%s\"'", variableName)
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "MariaDBReadOnlyVariableIsON")
	out, err := bash.RunCommand(cmd)
	if err != nil {
		c.Logger.Debug("failed to show variable", "name", variableName, "error", err)
		return false
	}

	s := string(out)
	return strings.Contains(s, "read_only") && strings.Contains(s, "ON")
}

// TurnOffBoolVariable implements Connector
func (c *MySQLCommandConnector) TurnOffBoolVariable(variableName string) error {
	cmd := fmt.Sprintf("mysql -e 'set global %s=0'", variableName)
	c.Logger.Info("execute command", "command", cmd, "callerFn", "TurnOffBoolVariable")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to set '%s' variable to 0: %w", variableName, err)
	}

	return nil
}

// TurnOnBoolVariable implements Connector
func (c *MySQLCommandConnector) TurnOnBoolVariable(variableName string) error {
	cmd := fmt.Sprintf("mysql -e 'set global %s=1'", variableName)
	c.Logger.Info("execute command", "command", cmd, "callerFn", "TurnOnBoolVariable")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to set '%s' variable to 1: %w", variableName, err)
	}

	return nil
}

// StopReplica implements Connector
func (c *MySQLCommandConnector) StartReplica() error {
	cmd := "mysql -e 'start replica'"
	c.Logger.Info("execute command", "command", cmd, "callerFn", "StartReplica")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to start replica: %w", err)
	}

	return nil
}

// StopReplica implements Connector
func (c *MySQLCommandConnector) StopReplica() error {
	cmd := "mysql -e 'stop replica'"
	c.Logger.Info("execute command", "command", cmd, "callerFn", "StopReplica")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to stop replica: %w", err)
	}

	return nil
}

// ResetAllReplicas implements Connector
func (c *MySQLCommandConnector) ResetAllReplicas() error {
	cmd := "mysql -e 'reset replica all'"
	c.Logger.Info("execute command", "command", cmd, "callerFn", "ResetAllReplicas")
	if _, err := bash.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to reset all replicas: %w", err)
	}

	return nil
}

// ChangeMasterTo implements Connector
func (c *MySQLCommandConnector) ChangeMasterTo(
	master MasterInstance,
) error {
	changeMasterOpts := []string{}
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_host = \"%s\"", master.Host))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_port = %d", master.Port))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_user = \"%s\"", master.User))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_password = \"%s\"", master.Password))
	changeMasterOpts = append(changeMasterOpts, fmt.Sprintf("master_use_gtid = %s", master.UseGTID))

	cmd := fmt.Sprintf("mysql -e 'change master to %s'", strings.Join(changeMasterOpts, ", "))
	c.Logger.Info("execute command", "command", cmd, "callerFn", "ChangeMasterTo")
	if out, err := bash.RunCommand(cmd); err != nil {
		c.Logger.Debug("changeMasterConnectConfig() output", "output", string(out))
		return fmt.Errorf("failed to change master connection config: %w", err)
	}

	return nil
}

// ShowReplicationStatus implements Connector
func (c *MySQLCommandConnector) ShowReplicationStatus() (ReplicationStatus, error) {
	cmd := "mysql -e 'show replica status\\G'"
	c.Logger.Debug("execute command", "command", cmd, "callerFn", "showReplicaStatusCommand")
	out, err := bash.RunCommand(cmd)
	if err != nil {
		return ReplicationStatus{}, fmt.Errorf("failed to show replica status: %w", err)
	}

	return parseShowReplicaStatusOutput(string(out)), nil
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
