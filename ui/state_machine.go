package ui

import (
	"context"
	"strconv"
	"time"

	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"
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
		Map to hold all instances, including drivers, groups & DBMS
	*/
	Instances map[string]types.Instance = make(map[string]types.Instance)

	/*
		Contains the key to all active connections that should be displayed
		[0] will be the current connection
	*/
	ActiveConns []string
	/*
		Keychain for ActiveConns popped keys
	*/
	IdleConns []string

	/*
		Tracking variable to limit input
	*/
	LastInputTime time.Time

	/*
		Map to hold all queries for ease of selection
		^ADD MAP FOR PLUGINS HERE
	*/
	GlobalQueryMap  map[types.DBMS_t]map[string]func() string = make(map[types.DBMS_t]map[string]func() string)
	MySQLQueries    map[string]func() string                  = make(map[string]func() string)
	PostgresQueries map[string]func() string                  = make(map[string]func() string)
)

func InterfaceLoop() {
	io.ReadInstances(Instances)

	/*
		If user doesn't specify arguments on run
		prompt connection details and put in []Instance and/or .conf
		Note: os.Args[0] == bin name, so args start @ index 1

		If user specifies correct number of arguments, attempt parse
		If parsing fail, error is thrown

		If user specifies wrong number of arguments, exit with code 1
	*/

	io.ReadArgs(Instances)
	io.SyncConfig(Instances)

	interval_int, _ := strconv.Atoi(io.FetchSetting("refresh-rate"))
	Interval = time.Duration(interval_int) * time.Millisecond

	/*
		Open & map all connections
		Set first connection as current
	*/
	if len(Instances) > 0 {
		for _, inst := range Instances {
			var key string = inst.ConnName
			var err error
			inst.Driver, err = queries.Connect(inst)
			Instances[key] = inst
			if err == nil {
				ActiveConns = append(ActiveConns, key)
			} else {
				IdleConns = append(IdleConns, key)
			}
		}
	}

	/*
		If no config found, force config state
		Else, startup-view
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
		case types.THREAD_ANALYSIS:
			DisplayThreadAnalysis()
			Laststate = types.THREAD_ANALYSIS
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
				if queries.Ping(inst) {
					inst.Driver.Close()
				}
			}
			return
		default:
			DisplayMenu()
			Laststate = types.MENU
			break
		}
	}
}

/*
Smart connection pooling system
*/
func connectionSanitiser(ctx context.Context, cancel context.CancelFunc) {
	var (
		ticker *time.Ticker = time.NewTicker(Interval * 10)
	)

	for {
		select {
		case <-ticker.C:
			/*
				Send back to config if no pingable connections
			*/
			if len(ActiveConns) == 0 && State != types.CONFIGS {
				State = types.CONFIGS
				cancel()
			}

			/*
				Clean up active connections
			*/
			for _, key := range ActiveConns {
				if !queries.Ping(Instances[key]) {
					ActiveConns = utility.PopString(ActiveConns, key)
					IdleConns = append(IdleConns, key)
				}
			}

			/*
				Update idle connections
			*/
			for _, key := range IdleConns {
				if queries.Ping(Instances[key]) {
					IdleConns = utility.PopString(IdleConns, key)
					queries.Connect(Instances[key])
					ActiveConns = append(ActiveConns, key)
				}
			}

		case <-ctx.Done():
			cancel()
		}
	}
}
