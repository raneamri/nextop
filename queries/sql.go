package queries

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

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
func Connect(instance types.Instance) (*sql.DB, error) {
	var (
		driver *sql.DB
		err    error
	)
	driver, err = sql.Open(utility.Strdbms(instance.DBMS), string(instance.DSN))
	if err != nil {
		driver.Close()
		return nil, err
	}
	driver.SetConnMaxLifetime(time.Minute * 3)
	driver.SetMaxOpenConns(10)
	driver.SetMaxIdleConns(10)
	if err := driver.Ping(); err != nil {
		driver.Close()
		return nil, err
	}
	return driver, nil
}

/*
Attempts to connect using a dsn
Returns true on success, false on fail
Used to authenticate connections
*/
func Ping(instance types.Instance) bool {
	var (
		driver *sql.DB
		err    error
	)
	driver, err = sql.Open(utility.Strdbms(instance.DBMS), string(instance.DSN))
	if err != nil {
		return false
	}
	if err := driver.Ping(); err != nil {
		return false
	}
	driver.Close()
	return true
}

/*
Queries a custom query
Works with any query
*/
func GetLongQuery(driver *sql.DB, query string) [][]string {
	rows, err := Query(driver, query)
	if err != nil {
		return nil
	}
	_, result, _ := GetData(rows)

	return result

}

/*
Curtesy of https://github.com/lefred
*/
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

/*
Performs a query given a connection and a statement
Helper function for other query-type functions
*/
func Query(db *sql.DB, stmt string) (*sql.Rows, error) {
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
