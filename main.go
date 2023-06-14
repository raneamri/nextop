package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/services"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/ui"
	"github.com/raneamri/gotop/utility"
)

/*
View concurrencies
View plugins
View make file
*/

func main() {

	utility.ClearTerminal()

	/*
		Start logging locks
	*/
	utility.StartQueue()

	/*
		Slice to store all instances
	*/
	var (
		instances []types.Instance
		cpool     []*sql.DB
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
		Fetch configs from args if any
	*/
	instances = io.ReadArgs(instances)

	/*
		Start connection pool
		Note: add further mapping to tie instance to connection
	*/
	for i, instance := range instances {
		if len(cpool) <= i || cpool[i] == nil {
			cpool = append(cpool, services.LaunchInstance(instance))
			if cpool[i] == nil {
				fmt.Println("Connection error.")
			}
		}
	}

	ui.InterfaceLoop(instances, cpool)
}

func Version() string {
	return "v0.0.0"
}
