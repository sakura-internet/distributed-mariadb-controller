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
