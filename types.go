package main

type dbms_t int

const (
	MYSQL dbms_t = iota
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
	ndx int

	/*
		Stores what database management tool instance is in
		This value is used as a key for command retrieval later
	*/
	dbms   dbms_t
	dbname string

	/*
		Login credentials
		Password stored securely
	*/
	user string
	pass []byte

	/*
		Required to access database
		Reminder: localhost is host := 127.0.0.1
	*/
	port int
	host int
}
