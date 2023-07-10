package types

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type (
	DBMS_t  int
	State_t int
)

/*
DBMS types
^ADD DBMS FOR PLUGIN HERE
*/
const (
	MYSQL DBMS_t = iota
	POSTGRE
)

/*
State machine tracking variables
^ADD STATE FOR PLUGIN HERE
*/
const (
	MENU State_t = iota
	PROCESSLIST
	DB_DASHBOARD
	MEM_DASHBOARD
	ERR_LOG
	LOCK_LOG
	REPLICATION
	CONFIGS
	QUIT
)

/*
Holds all data of an instance and its database
Ideally, it will be maleable to most/all DBMSs
as of now, matches MySQL syntax
*/
type Instance struct {
	/*
		Stores what database management tool instance is in
		This value is used as a key for command retrieval later
	*/
	DBMS DBMS_t

	/*
		DSN, stored as byte slice to protect password
	*/
	DSN []byte

	/*
		SQL driver
	*/
	Driver *sql.DB
	Online bool

	/*
		Manually assigned connection name
	*/
	ConnName string

	/*
		Group name
	*/
	Group string
}
