package io

import (
	"fmt"
	"os"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
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
	fmt.Println("DSN: ")
	fmt.Scanf("%s", &newInstance.DSN)

	utility.ClearTerminal()
	fmt.Println("Enter connection name: ")
	fmt.Scanf("%s", &newInstance.ConnName)
	if newInstance.ConnName == "" {
		newInstance.ConnName = "<unnamed>"
	}

	utility.ClearTerminal()
	return newInstance
}

func NoArgStartup(instances []types.Instance) []types.Instance {
	var loadnew string
	/*
		Prompt user if they want to use configged only if
		config isn't empty
	*/
	if len(instances) != 0 {
		fmt.Println("Load in with new instance?: [yes] ")
		fmt.Scanf("%s", &loadnew)
	} else {
		loadnew = "YES"
	}
	if utility.Fstr(loadnew) == "YES" || utility.Fstr(loadnew) == "Y" {
		inst := NewInstance()
		/*
			Write to config conditionally
		*/
		var write string
		fmt.Printf("Write to config?: [yes] ")
		fmt.Scanf("%s", &write)
		if utility.Fstr(write) == "YES" {
			/*
				Store config in dynamic slice
			*/
			instances = utility.PushInstance(instances, inst)
		}
		return instances
	} else {
		return instances
	}
}
