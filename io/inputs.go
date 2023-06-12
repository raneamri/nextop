package io

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

	"golang.org/x/term"
)

/*
Prompts user to pick a DBMS
*/
func PickDBMS() types.DBMS_t {
	var input string
	fmt.Println("DBMS: ")
	fmt.Scanf("%s", &input)
	input = utility.Fstr(input)

	if input == "MYSQL" {
		return types.MYSQL
	} else if input == "ORACLE" {
		return types.ORACLE
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
func NewInstance() types.Instance {
	var newInstance types.Instance
	newInstance.DBMS = PickDBMS()

	utility.ClearTerminal()
	fmt.Println("Enter username: ")
	fmt.Scanf("%s", &newInstance.User)

	utility.ClearTerminal()
	fmt.Print("Enter password: \n")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	newInstance.Pass = password

	utility.ClearTerminal()
	fmt.Println("Enter port (default:3306): ")
	var portIn string
	fmt.Scanf("%d", &portIn)
	if portIn == "" {
		newInstance.Port = 3306
	} else {
		newInstance.Port, _ = strconv.Atoi(portIn)
	}

	utility.ClearTerminal()
	fmt.Println("Enter host (default:127.0.0.1): ")
	fmt.Scanf("%s", &newInstance.Host)
	if newInstance.Host == "" {
		newInstance.Host = "127.0.0.1"
	}

	utility.ClearTerminal()
	fmt.Println("Enter database name (default:none): ")
	fmt.Scanf("%s", &newInstance.Dbname)
	if newInstance.Dbname == "" {
		newInstance.Dbname = "none"
	}

	utility.ClearTerminal()
	return newInstance
}
