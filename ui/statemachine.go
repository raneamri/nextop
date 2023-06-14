package ui

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func InterfaceLoop(instances []types.Instance, cpool []*sql.DB) {
	/*
		Interface parameters
	*/
	var (
		state    types.State_t = types.DB_DASHBOARD
		fps      int           = 60
		interval time.Duration = time.Duration(fps/60) * time.Second
	)

	for 1 == 1 {
		utility.ClearTerminal()
		switch state {
		case types.MENU:

			break
		case types.PROCESSLIST:

			break
		case types.DB_DASHBOARD:
			dashboard := InitDashboard(instances, cpool)
			fmt.Println(dashboard.String())

			break
		case types.MEM_DASHBOARD:

			break
		case types.ERR_LOG:

			break
		case types.LOCK_LOG:
			/*
				Assuming this shows current locks in backend
			*/

			break
		case types.CONFIGS:
			/*
				Force user to this state if no configs are found and if launched w/o args.
				Prompt user to connect to database
			*/

			break
		case types.HELP:
			/*
				Display help text and GitHub
			*/

			break
		case types.QUIT:
			/*
				Perform cleanup and clean program
			*/
			for _, conn := range cpool {
				conn.Close()
			}
			os.Exit(1)
		}
		time.Sleep(interval)
	}
}
