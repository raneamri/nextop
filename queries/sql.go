package queries

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

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

	switch instance.DBMS {
	case types.MYSQL:
		driver.SetConnMaxLifetime(time.Minute * 3)
		driver.SetMaxOpenConns(10)
		driver.SetMaxIdleConns(10)
	case types.POSTGRES:
		//...
	}

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
