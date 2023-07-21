package ui

import (
	"strconv"
	"time"

	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"
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
		Key is made of connection name, an indicative character to show there is a group and the group name
	*/
	Instances map[string]types.Instance = make(map[string]types.Instance)

	/*
		Contains the key to all active connections that should be displayed
	*/
	ActiveConns []string
	/*
		Keychain for ActiveConns popped keys
	*/
	IdleConns []string
	/*
		Holds key to main connection to be displayed
		since some displays aren't big enough to show all instances
	*/
	CurrConn string

	/*
		Tracking variable to limit input
	*/
	LastInputTime time.Time

	/*
		Map to hold all queries for ease of selection
		^ADD MAP FOR PLUGINS HERE
	*/
	GlobalQueryMap map[types.DBMS_t]map[string]func() string = make(map[types.DBMS_t]map[string]func() string)
	MySQLQueries   map[string]func() string                  = make(map[string]func() string)
)

func InterfaceLoop() {
	/*
		Attempt to fetch config from .conf
	*/
	io.ReadInstances(Instances)

	/*
		If user doesn't specify arguments on run
		prompt connection details and put in []Instance and/or .conf
		Note: os.Args[0] == bin name, so args start @ index 1

		If user specifies correct number of arguments, attempt parse
		If parsing fail, error is thrown

		If user specifies wrong number of arguments, exit with code 1
	*/
	if len(Instances) == 0 {
		io.ReadArgs(Instances)
	}

	/*
		Syncs dynamically stored configs to statically stored configs
		Syncing involves writing to config (view files.go)
	*/
	io.SyncConfig(Instances)

	/*
		Fetch refresh rate from config
	*/
	interval_int, _ := strconv.Atoi(io.FetchSetting("refresh-rate"))
	Interval = time.Duration(interval_int) * time.Millisecond
	err_interval_int, _ := strconv.Atoi(io.FetchSetting("errlog-refresh-rate"))
	ErrInterval = time.Duration(err_interval_int) * time.Millisecond

	/*
		Open & map all connections
		Set first connection as current
	*/
	var flag bool = true
	if len(Instances) > 0 {
		for _, inst := range Instances {
			var key string = inst.ConnName
			if flag {
				CurrConn = key
				flag = false
			}
			var err error
			inst.Driver, err = queries.Connect(inst)
			Instances[key] = inst
			if err == nil {
				ActiveConns = append(ActiveConns, key)
			}
		}
	}

	/*
		If no config found, force config state
		Else, processlist
	*/
	if len(ActiveConns) == 0 {
		State = types.CONFIGS
	} else {
		State = utility.Statestr(io.FetchSetting("startup-view"))
	}

	/*
		Set up query maps
	*/
	queries.MapMySQL(MySQLQueries)
	GlobalQueryMap[types.MYSQL] = MySQLQueries

	for true {
		switch State {
		case types.MENU:
			DisplayMenu()
			Laststate = types.MENU
			break
		case types.PROCESSLIST:
			DisplayProcesslist()
			Laststate = types.PROCESSLIST
			break
		case types.DB_DASHBOARD:
			DisplayInnoDbDashboard()
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
		case types.REPLICATION:
			DisplayReplication()
			Laststate = types.REPLICATION
			break
		case types.TRANSACTIONS:
			DisplayTransactions()
			Laststate = types.TRANSACTIONS
			break
		case types.CONFIGS:
			io.SyncConfig(Instances)
			DisplayConfigs()
			Laststate = types.CONFIGS
			break
		case types.QUIT:
			/*
				Perform cleanup and close program
			*/
			for _, inst := range Instances {
				inst.Driver.Close()
			}
			return
		default:
			DisplayMenu()
			Laststate = types.MENU
			break
		}
	}
}
