package main

import (
	"fmt"
	"time"

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
		Slice to store all instances
	*/
	var (
		instances []types.Instance
		err       error
	)
	/*
		Attempt to fetch config from .conf
	*/
	instances, err = io.ReadConfig()
	if err != nil {
		errors.CatchConfigReadError(err, instances)
	}
	/*
		Fetch configs from prompt/args
	*/
	io.ReadArgs(instances)
	/*
		Merge .conf & input
	*/

	/*
		Temporary system in the form of "game loop"
	*/
	var (
		fps      int
		interval time.Duration
	)
	fps = 60
	interval = time.Duration(fps/60) * time.Second
	for _, inst := range instances {
		inst.Driver = services.LaunchInstance(inst)
	}
	for 1 == 1 {
		utility.ClearTerminal()
		dashboard := ui.InitDashboard(instances)
		fmt.Println(dashboard.String())
		time.Sleep(interval)
	}

}

func Version() string {
	return "v0.0.0"
}
