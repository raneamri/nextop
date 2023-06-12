package types

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DBMS_t int

const (
	MYSQL DBMS_t = iota
	ORACLE
)

/*
Holds all data of an instance and its database
Ideally, it will be maleable to most/all DBMSs
as of now, matches MySQL syntax
*/
type Instance struct {
	/*
		Stores index of instance
		Note: hard code max
	*/
	Ndx int

	/*
		Stores what database management tool instance is in
		This value is used as a key for command retrieval later
	*/
	DBMS DBMS_t

	/*
		Login credentials
		Password stored securely
		Note: find solution to store pass in config
	*/
	User string
	Pass []byte

	/*
		Required to access database
		defaults to {
			port 3306
			host 127.0.0.1
			dbname ndx
		}
		Reminder: localhost is host := 127.0.0.1
	*/
	Port int
	Host string

	Dbname string

	/*
		Database handle
	*/
	DB *sql.DB
}
