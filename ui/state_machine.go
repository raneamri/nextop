package ui

import (
	"database/sql"
	"time"

	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/raneamri/gotop/types"
)

var (
	State     types.State_t
	Laststate types.State_t
	Interval  time.Duration = 500 * time.Millisecond
	PoolNdx   int           = 0
)

func InterfaceLoop(instances []types.Instance, cpool []*sql.DB) {
	/*
		Interface parameters
		Open a tcell for the interface
	*/
	t, err := tcell.New()
	defer t.Close()
	if err != nil {
		panic(err)
	}

	/*
		Decide initial state
		If no config found, force config state
		Else, menu state
		Note: change else to menu after testing is done
	*/
	if len(instances) == 0 {
		State = types.CONFIGS
	} else {
		State = types.MENU
	}

	for true {
		switch State {
		case types.MENU:
			DrawMenu(t)
			Laststate = types.MENU
			break
		case types.PROCESSLIST:
			DisplayProcesslist(t, cpool)
			Laststate = types.PROCESSLIST
			if State == types.PROCESSLIST {
				State = types.MENU
			}
			break
		case types.DB_DASHBOARD:
			Laststate = types.DB_DASHBOARD
			DisplayDbDashboard(t, cpool)
			if State == types.DB_DASHBOARD {
				State = types.MENU
			}
			break
		case types.MEM_DASHBOARD:
			Laststate = types.MEM_DASHBOARD
			DisplayMemory(t, cpool)
			if State == types.MEM_DASHBOARD {
				State = types.MENU
			}
			break
		case types.ERR_LOG:
			DisplayErrorLog(t, cpool)
			Laststate = types.ERR_LOG
			if State == types.ERR_LOG {
				State = types.MENU
			}
			break
		case types.LOCK_LOG:
			DisplayLocks(t, cpool)
			Laststate = types.LOCK_LOG
			if State == types.LOCK_LOG {
				State = types.MENU
			}
			break
		case types.CONFIGS:
			/*
				Force user to this state if no configs are found and if launched w/o args.
				Prompt user to connect to database
			*/
			DisplayConfigs(t, instances, cpool)
			Laststate = types.CONFIGS
			if State == types.CONFIGS {
				State = types.MENU
			}
			break
		case types.HELP:
			/*
				Display help text and GitHub
			*/
			DrawHelp(t)
			Laststate = types.HELP
			if State == types.HELP {
				State = types.MENU
			}
			break
		case types.QUIT:
			/*
				Perform cleanup and close program
			*/
			//t.Close() fixes input buffer overflow ?
			for _, conn := range cpool {
				conn.Close()
			}
			return
		}
	}
}
