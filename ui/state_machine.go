package ui

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/raneamri/gotop/db"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/types"
)

var (
	/*
		State trackers
	*/
	State     types.State_t
	Laststate types.State_t
	/*
		Refresh rate
	*/
	Interval time.Duration
	/*
		Holds all drivers
	*/
	ConnPool map[string]*sql.DB = make(map[string]*sql.DB)
	/*
		Contains the key to all active connections that should be displayed
	*/
	ActiveConns []string
	/*
		Keychain for ActiveConns popped keys
	*/
	InactiveConns []string
	/*
		Holds key to main connection to be displayed
		since some displays aren't big enough to show all instances
	*/
	CurrConn string
)

func InterfaceLoop(instances []types.Instance) {
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
		Fetch refresh rate from config
	*/
	interval_int, _ := strconv.Atoi(io.FetchSetting("refresh-rate"))
	Interval = time.Duration(interval_int) * time.Millisecond

	/*
		Open & map all connections
	*/
	if len(instances) > 0 {
		for i, inst := range instances {
			if i == 0 {
				CurrConn = inst.ConnName
			}
			ConnPool[inst.ConnName] = db.Connect(inst)
			ActiveConns = append(ActiveConns, inst.ConnName)
		}
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
			DisplayProcesslist(t)
			Laststate = types.PROCESSLIST
			if State == types.PROCESSLIST {
				State = types.MENU
			}
			break
		case types.DB_DASHBOARD:
			Laststate = types.DB_DASHBOARD
			DisplayDbDashboard(t)
			if State == types.DB_DASHBOARD {
				State = types.MENU
			}
			break
		case types.MEM_DASHBOARD:
			Laststate = types.MEM_DASHBOARD
			DisplayMemory(t)
			if State == types.MEM_DASHBOARD {
				State = types.MENU
			}
			break
		case types.ERR_LOG:
			DisplayErrorLog(t)
			Laststate = types.ERR_LOG
			if State == types.ERR_LOG {
				State = types.MENU
			}
			break
		case types.LOCK_LOG:
			DisplayLocks(t)
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
			DisplayConfigs(t, instances)
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
			for _, key := range ActiveConns {
				ConnPool[key].Close()
			}
			return
		}
	}
}
