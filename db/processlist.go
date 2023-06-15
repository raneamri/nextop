package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func GetProcesslist(driver *sql.DB) ([]string, [][]string, error) {
	statement := `select pps.PROCESSLIST_COMMAND AS command,
                                  pps.THREAD_ID AS thd_id, pps.PROCESSLIST_ID AS conn_id,
                                  conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state,
                                  if((pps.NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection')),
                                   concat(pps.PROCESSLIST_USER,'@',pps.PROCESSLIST_HOST),
                                   replace(pps.NAME,'thread/','')) AS user,
                                  pps.PROCESSLIST_DB AS db, sys.format_statement(pps.PROCESSLIST_INFO) AS current_statement,
                                  if(isnull(esc.END_EVENT_ID), format_pico_time(esc.TIMER_WAIT),NULL) AS statement_latency,
                                  format_pico_time(esc.LOCK_TIME) AS lock_latency,
                                  if(isnull(esc.END_EVENT_ID),esc.TIMER_WAIT,0) AS sort_time
                            from (performance_schema.threads pps
                            left join performance_schema.events_statements_current esc
                                on (pps.THREAD_ID = esc.THREAD_ID))
							left join performance_schema.session_connect_attrs conattr_pid
        						 on((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID) and (conattr_pid.ATTR_NAME = '_pid'))
                            where pps.PROCESSLIST_ID is not null
                              and pps.PROCESSLIST_COMMAND <> 'Daemon'
                            order by sort_time desc
                        `
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
