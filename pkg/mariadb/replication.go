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
