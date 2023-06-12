package services

import (
	"database/sql"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/raneamri/gotop/types"

	_ "github.com/go-sql-driver/mysql"
)

/*
Note: add SQL to function names once development for over DBMS starts
*/

/*
	By default, they evaluate to {
			ConnMaxLifetime time.Minute * 3
			SetMaxOpenConns 10
			SetMaxIdleConns  "
	}

Quoting https://github.com/go-sql-driver/mysql#features:

	".SetConnMaxLifetime() is required to ensure connections are closed by the driver
	safely before connection is closed by MySQL server, OS, or other middlewares. Since
	some middlewares close idle connections by 5 minutes, we recommend timeout shorter
	than 5 minutes. This setting helps load balancing and changing system variables too.

	.SetMaxOpenConns() is highly recommended to limit the number of connection used by
	the application. There is no recommended limit number because it depends on application
	and MySQL server.

	.SetMaxIdleConns() is recommended to be set same to .SetMaxOpenConns(). When it is
	smaller than SetMaxOpenConns(), connections can be opened and closed much more frequently
	than you expect. Idle connections can be closed by the .SetConnMaxLifetime(). If you
	want to close idle connections more rapidly, you can use .SetConnMaxIdleTime() since Go 1.15.
	"
*/
func SetParameters(db *sql.DB) {
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
}

/*
Unwrap instance into db pointer
*/
func LaunchInstance(instance types.Instance) *sql.DB {
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	if instance.Driver == nil {
		driver, err := sql.Open("mysql", instance.User+":"+string(instance.Pass)+"@tcp("+fmt.Sprint(instance.Host)+":"+fmt.Sprint(instance.Port)+")/"+instance.Dbname)
		if err != nil || driver == nil {
			fmt.Println("Improper db connection. View:")
			panic(err)
		}
		return driver
	}
	return instance.Driver
}

/*
Retrieves uptime of database and returns it formatted.
*/
func GetUptime(db *sql.DB) float64 {
	if db == nil {
		return -1
	}
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	var (
		uptime  float64
		discard string
	)
	query := "SHOW GLOBAL STATUS LIKE 'Uptime'"
	if err := db.QueryRow(query).Scan(&discard, &uptime); err != nil {
		panic(err)
	}
	return uptime
}

func GetQPS(db *sql.DB) float64 {
	if db == nil {
		return -1
	}
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()
	var (
		queries,
		uptime,
		qps float64
		discard string
	)
	query := "SHOW GLOBAL STATUS LIKE 'Queries'"

	/*
		Retrieve 'Queries' values from 'SHOW GLOBAL STATUS'
		Retrieve 'Uptime' with similar method from getUptime
	*/
	if err := db.QueryRow(query).Scan(&discard, &queries); err != nil {
		panic(err)
	}

	/*
		Calculate QPS from uptime
		Decrease queries by 2 to account for the queries required
		to calculate QPS
	*/
	uptime = GetUptime(db)
	queries -= 2
	if uptime > 0 {
		qps = math.Round(queries / uptime)
	}

	return qps
}
