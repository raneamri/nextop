package db

import (
	"database/sql"
)

func GetInnodbInfo(driver *sql.DB) ([]string, [][]string, error) {
	statement := `SHOW ENGINE INNODB STATUS`

	rows, err := Query(driver, statement)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := GetData(rows)
	if err != nil {
		return nil, nil, err
	}

	return cols, data, nil
}
