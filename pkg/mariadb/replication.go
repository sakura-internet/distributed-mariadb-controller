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

type MasterInstance struct {
	Host     string
	Port     int
	User     string
	Password string
	UseGTID  MasterUseGTIDValue
}

type MasterUseGTIDValue string

const (
	MasterUseGTIDValueCurrentPos MasterUseGTIDValue = "current_pos"
	MasterUseGTIDValueSlavePos   MasterUseGTIDValue = "slave_pos"
	MasterUseGTIDValueNo         MasterUseGTIDValue = "no"
)

type ReplicationStatus map[string]string

const (
	ReplicationStatusSlaveIORunning     = "Slave_IO_Running"
	ReplicationStatusSlaveSQLRunning    = "Slave_SQL_Running"
	ReplicationStatusReadMasterLogPos   = "Read_Master_Log_Pos"
	ReplicationStatusRelayMasterLogFile = "Relay_Master_Log_File"
	ReplicationStatusMasterLogFile      = "Master_Log_File"
	ReplicationStatusExecMasterLogPos   = "Exec_Master_Log_Pos"

	ReplicationStatusSlaveIORunningYes  = "Yes"
	ReplicationStatusSlaveSQLRunningYes = "Yes"
)
