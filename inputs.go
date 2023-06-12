package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

/*
Prompts user to pick a DBMS
*/
func PickDBMS() DBMS_t {
	var input string
	fmt.Println("DBMS: ")
	fmt.Scanf("%s", &input)
	input = Fstr(input)

	if input == "MYSQL" {
		return MYSQL
	} else if input == "ORACLE" {
		return ORACLE
	}

	fmt.Println("Unaccomodated/non-existant DBMS.")
	/*
		Note: switch for loop
	*/
	os.Exit(1)
	return -1
}

/*
Creates a new instance and returns it
*/
func newInstance() Instance {
	var newInstance Instance
	newInstance.DBMS = PickDBMS()

	ClearTerminal()
	fmt.Println("Enter username: ")
	fmt.Scanf("%s", &newInstance.User)

	ClearTerminal()
	fmt.Print("Enter password: \n")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	newInstance.Pass = password

	ClearTerminal()
	fmt.Println("Enter port (default:3306): ")
	fmt.Scanf("%d", &newInstance.Port)
	if fmt.Sprint(newInstance.Port) == "" {
		newInstance.Port = 3306
	}

	ClearTerminal()
	fmt.Println("Enter host (default:127.0.0.1): ")
	fmt.Scanf("%s", &newInstance.Host)
	if newInstance.Host == "" {
		newInstance.Host = "127.0.0.1"
	}

	ClearTerminal()
	fmt.Println("Enter database name (default:none): ")
	fmt.Scanf("%s", &newInstance.Dbname)
	if newInstance.Dbname == "" {
		newInstance.Dbname = "none"
	}

	ClearTerminal()
	return newInstance
}
