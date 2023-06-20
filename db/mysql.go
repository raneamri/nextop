package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

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
/*
Unwrap instance into db pointer
*/
func Connect(instance types.Instance) *sql.DB {
	driver, _ := sql.Open(utility.Strdbms(instance.DBMS), string(instance.DSN))
	driver.SetConnMaxLifetime(time.Minute * 3)
	driver.SetMaxOpenConns(10)
	driver.SetMaxIdleConns(10)
	if err := driver.Ping(); err != nil {
		driver.Close()
		panic(err)
	}
	return driver
}

/*
Attempts to connect using a dsn
Returns true on success, false on fail
Used to authentificate connections
*/
func Ping(instance types.Instance) bool {
	driver, _ := sql.Open(utility.Strdbms(instance.DBMS), string(instance.DSN))
	if err := driver.Ping(); err != nil {
		driver.Close()
		return false
	}
	driver.Close()
	return true
}

/*
Establishes a connection to that database and finds the specified status
*/
func GetStatus(driver *sql.DB, parameters []string) []string {
	var results []string

	for _, param := range parameters {
		query := `SHOW STATUS LIKE '` + param + `';`
		rows, _ := Query(driver, query)
		_, result, _ := GetData(rows)
		results = append(results, result[0][1])
	}

	return results
}

/*
Establishes a connection to that database and finds the specified variable
*/
func GetVariable(driver *sql.DB, parameters []string) []string {
	var results []string

	for _, param := range parameters {
		query := `SHOW VARIABLE LIKE '` + param + `';`
		rows, _ := Query(driver, query)
		_, result, _ := GetData(rows)
		results = append(results, result[0][2])
	}

	return results
}

/*
Retrieves uptime of database and returns it formatted.
Note: adjust methods
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

func GetThreads(db *sql.DB) int {
	if db == nil {
		return -1
	}
	var mutex sync.Mutex
	mutex.Lock()
	defer mutex.Unlock()

	var (
		threads int
		discard string
	)

	query := "SHOW GLOBAL STATUS LIKE 'Threads_connected'"

	if err := db.QueryRow(query).Scan(&discard, &threads); err != nil {
		panic(err)
	}

	return threads
}

func GetData(rows *sql.Rows) ([]string, [][]string, error) {
	var result [][]string
	defer rows.Close()

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.RawBytes)
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, nil, err
		}
		var resultRow []string
		for i, col := range vals {
			var value string
			if col == nil {
				value = "NULL"
			} else {
				switch colTypes[i].DatabaseTypeName() {
				case "VARCHAR", "CHAR", "TEXT":
					value = fmt.Sprintf("%s", col)
				case "BIGINT":
					value = fmt.Sprintf("%s", col)
				case "INT":
					value = fmt.Sprintf("%d", col)
				case "DECIMAL":
					value = fmt.Sprintf("%s", col)
				default:
					value = fmt.Sprintf("%s", col)
				}
			}
			value = strings.Replace(value, "&", "", 1)
			resultRow = append(resultRow, value)
		}
		result = append(result, resultRow)
	}
	return cols, result, nil
}

func DisplayData(cols []string, result [][]string) {
	// Print column names
	for _, col := range cols {
		fmt.Printf("%s\t", col)
	}
	fmt.Println()

	// Print data rows
	for _, row := range result {
		for _, val := range row {
			fmt.Printf("%s\t", val)
		}
		fmt.Println()
	}
}

/*
Performs a query given a connection and a statement
*/
func Query(db *sql.DB, stmt string) (*sql.Rows, error) {
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
