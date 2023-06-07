package main

import (
	"fmt"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/term"
)

func main() {

	clearTerminal()

	/*
		Slice to store all instances
	*/
	var instances []Instance

	/*
		User prompted to enter what DBMS they're using
	*/
	inst := newInstance(pickDBMS())
	instances = push_instance(instances, inst)
}

/*
Prompts user to book a DBMS
*/
func pickDBMS() dbms_t {
	var input string
	fmt.Println("DBMS: ")
	fmt.Scanf("%s", &input)
	input = fstr(input)

	if input == "MYSQL" {
		return MYSQL
	} else if input == "ORACLE" {
		return ORACLE
	}

	/*
		Remove warning
		Should never be reached
	*/
	return -1
}

/*
Creates a new instance profile and returns it
*/
func newInstance(dbms dbms_t) Instance {
	var newInstance Instance
	newInstance.dbms = dbms

	clearTerminal()
	fmt.Println("Enter database name: ")
	fmt.Scanf("%s", &newInstance.user)

	clearTerminal()
	fmt.Println("Enter username: ")
	fmt.Scanf("%s", &newInstance.user)

	clearTerminal()
	fmt.Print("Enter password: \n")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	newInstance.pass = password

	clearTerminal()
	fmt.Println("Enter port: ")
	fmt.Scanf("%d", &newInstance.port)

	clearTerminal()
	fmt.Println("Enter host: ")
	fmt.Scanf("%d", &newInstance.host)

	return newInstance
}

/*
Pushes an item to the top of a slice
Also provides item with index
*/
func push_instance(instances []Instance, pushing Instance) []Instance {
	pushing.ndx = len(instances)
	instances = append(instances, pushing)
	return instances
}

func pop_instance(instances []Instance, popping Instance) []Instance {
	var rm int = -1
	for i, it := range instances {
		if it.dbms == popping.dbms && it.host == popping.host && it.port == popping.port && it.user == popping.user {
			rm = i
		}
	}

	if rm != -1 {
		// Create a new slice with one less element
		newinsts := make([]Instance, len(instances)-1)
		copy(newinsts[:rm], instances[:rm])
		copy(newinsts[rm:], instances[rm+1:])

		instances = newinsts
	} else {
		fmt.Println("Instance to pop unfound.")
	}

	/*
		Re-index instances
	*/
	for i, it := range instances {
		it.ndx = i
	}

	return instances
}
