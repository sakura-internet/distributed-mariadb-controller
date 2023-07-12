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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseShowReplicaStatusOutput(t *testing.T) {
	const input = `*************************** 1. row ***************************
                Slave_IO_State:
                   Master_Host: 8.8.8.8
                   Master_User: repl
                   Master_Port: 3306
                 Connect_Retry: 60
               Master_Log_File:
           Read_Master_Log_Pos: 4
                Relay_Log_File: relay-bin.000001
                 Relay_Log_Pos: 4
         Relay_Master_Log_File:
              Slave_IO_Running: No
             Slave_SQL_Running: Yes`

	result := parseShowReplicaStatusOutput(input)

	assert.Equal(t, "8.8.8.8", result["Master_Host"])
	assert.Equal(t, "No", result["Slave_IO_Running"])
	assert.Equal(t, "Yes", result["Slave_SQL_Running"])

}
