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

/*
Standard state machine to alter program states
Also tracks previous state
I believe it's guilty for irregular flash page buffering
*/

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
		Specific refresh rate for the error log
		Can be adjusted in config
		If you believe you will need regular and rapid filter changes, set to higher value
	*/
	ErrInterval time.Duration
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
	err_interval_int, _ := strconv.Atoi(io.FetchSetting("errlog-refresh-rate"))
	ErrInterval = time.Duration(err_interval_int) * time.Millisecond

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
		If no config found, force config state
		Else, menu state
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
			break
		case types.DB_DASHBOARD:
			Laststate = types.DB_DASHBOARD
			DisplayDbDashboard(t)
			break
		case types.MEM_DASHBOARD:
			Laststate = types.MEM_DASHBOARD
			DisplayMemory(t)
			break
		case types.ERR_LOG:
			DisplayErrorLog(t)
			Laststate = types.ERR_LOG
			break
		case types.LOCK_LOG:
			DisplayLocks(t)
			Laststate = types.LOCK_LOG
			break
		case types.CONFIGS:
			/*
				Force user to this state if no configs are found and if launched w/o args.
				Prompt user to connect to database
			*/
			DisplayConfigs(t, instances)
			Laststate = types.CONFIGS
			break
		case types.HELP:
			/*
				Display help text and repo
			*/
			DrawHelp(t)
			Laststate = types.HELP
			break
		case types.QUIT:
			/*
				Perform cleanup and close program
			*/
			t.Close()
			for _, key := range ActiveConns {
				ConnPool[key].Close()
			}
			return
		}
	}
}
