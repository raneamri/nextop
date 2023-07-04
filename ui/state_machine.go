package ui

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/raneamri/nextop/db"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"
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
		Map to hold all instances, including drivers, groups & DBMS
		Key is connection name
	*/
	Instances map[string]types.Instance = make(map[string]types.Instance)

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
			var key string = inst.ConnName
			if inst.Group != "" {
				key += "&" + inst.Group
			}
			if i == 0 {
				CurrConn = key
			}
			ConnPool[key] = db.Connect(inst)
			ActiveConns = append(ActiveConns, key)
		}
	}

	/*
		If no config found, force config state
		Else, menu state
	*/
	if len(instances) == 0 {
		State = types.CONFIGS
	} else {
		State = types.PROCESSLIST
	}

	for true {
		switch State {
		case types.MENU:
			DrawMenu()
			Laststate = types.MENU
			break
		case types.PROCESSLIST:
			DisplayProcesslist()
			Laststate = types.PROCESSLIST
			break
		case types.DB_DASHBOARD:
			DisplayDbDashboard()
			Laststate = types.DB_DASHBOARD
			break
		case types.MEM_DASHBOARD:
			DisplayMemory()
			Laststate = types.MEM_DASHBOARD
			break
		case types.ERR_LOG:
			DisplayErrorLog()
			Laststate = types.ERR_LOG
			break
		case types.LOCK_LOG:
			DisplayLocks()
			Laststate = types.LOCK_LOG
			break
		case types.CONFIGS:
			DisplayConfigs(instances)
			break
		case types.QUIT:
			/*
				Perform cleanup and close program
			*/
			for _, key := range ActiveConns {
				ConnPool[key].Close()
			}
			return
		}
	}
}
