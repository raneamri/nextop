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
	driver, err = sql.Open(utility.Strdbms(instance.DBMS), string(instance.DSN)+"?maxAllowedPacket=67108864")
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
Curtesy of https://github.com/lefred (altered method)
*/
func GetData(rows *sql.Rows) ([]string, [][]string, error) {
	var result [][]string
	defer rows.Close()

	_, err := rows.ColumnTypes()
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
		for _, col := range vals {
			var value string
			if col == nil {
				value = "NULL"
			} else {
				value = fmt.Sprintf("%s", col)
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
func Query(db *sql.DB, stmt string) (types.Query, error) {
	var (
		query types.Query
		rows  *sql.Rows
		err   error

		headers []string
		result  [][]string
	)

	rows, err = db.Query(stmt)

	if err != nil {
		return query, err
	}
	headers, result, err = GetData(rows)

	query.VarNames = headers
	query.RawData = result
	if err != nil {
		return query, err
	}
	return query, nil
}

func PromoteConnection(active_pool *[]string, idle_pool *[]string, conn string, instances map[string]types.Instance) {
	*idle_pool = utility.PopString(*idle_pool, conn)
	*active_pool = append(*active_pool, conn)

	/*
		Re-open connection to avoid a null dereference
	*/
	var inst types.Instance = types.Instance{
		DBMS:     instances[conn].DBMS,
		DSN:      instances[conn].DSN,
		ConnName: instances[conn].ConnName,
		Group:    instances[conn].Group,
	}

	inst.Driver, _ = Connect(inst)
	instances[conn] = inst
}

func DemoteConnection(active_pool *[]string, idle_pool *[]string, conn string) {
	*active_pool = utility.PopString(*active_pool, conn)
	*idle_pool = append(*idle_pool, conn)
}
