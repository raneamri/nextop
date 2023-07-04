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
*/
const (
	MYSQL DBMS_t = iota
	ORACLE
	POSTGRE
)

/*
State machine tracking variables
*/
const (
	MENU State_t = iota
	PROCESSLIST
	DB_DASHBOARD
	MEM_DASHBOARD
	ERR_LOG
	LOCK_LOG
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
