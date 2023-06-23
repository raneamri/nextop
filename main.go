package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/ui"
	"github.com/raneamri/gotop/utility"
)

/*
View concurrencies
View plugins
View make file
Get queries
Draw all interfaces
Make dynamic
Add filter & search
*/

func main() {

	utility.ClearTerminal()

	/*
		Slice to store all instances & connection pool
	*/
	var (
		instances []types.Instance
		err       error
	)
	/*
		Attempt to fetch config from .conf
	*/
	instances, err = io.ReadConfig(instances)
	if err != nil {
		errors.CatchConfigReadError(err, instances)
	}

	/*
		If user doesn't specify arguments on run
		prompt connection details and put in []Instance and/or .conf
		Note: os.Args[0] == binary, so args start @ index 1

		If user specifies correct number of arguments, attempt parse
		If parsing fail, error is thrown

		If user specifies wrong number of arguments, exit with code 1
	*/
	if len(os.Args) > 2 {
		instances = io.ReadArgs(instances)
	}

	/*
		Syncs dynamically stored configs to statically stored configs
		Syncing involves writing to config (view files.go)
	*/
	instances = io.SyncConfig(instances)

	utility.ClearTerminal()
	ui.InterfaceLoop(instances)
	fmt.Println("Bye")
}

func Version() string {
	return "v0.0.0"
}
