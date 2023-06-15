package ui

import (
	"database/sql"

	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/raneamri/gotop/types"
)

var (
	State     types.State_t
	Laststate types.State_t
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
			Laststate = types.PROCESSLIST
			break
		case types.DB_DASHBOARD:
			Laststate = types.DB_DASHBOARD
			break
		case types.MEM_DASHBOARD:
			Laststate = types.MEM_DASHBOARD
			break
		case types.ERR_LOG:
			Laststate = types.ERR_LOG
			break
		case types.LOCK_LOG:
			/*
				Assuming this shows current locks in backend
			*/
			Laststate = types.LOCK_LOG
			break
		case types.CONFIGS:
			/*
				Force user to this state if no configs are found and if launched w/o args.
				Prompt user to connect to database
			*/
			Laststate = types.CONFIGS
			break
		case types.HELP:
			/*
				Display help text and GitHub
			*/
			DrawHelp(t)

			State = Laststate
			Laststate = types.HELP
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
