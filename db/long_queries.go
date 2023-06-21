package db

func ProcesslistLongQuery() string {
	return `SELECT pps.PROCESSLIST_COMMAND AS command,
				pps.THREAD_ID AS thd_id, pps.PROCESSLIST_ID AS conn_id,
			conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state,
			if((pps.NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection')),
			concat(pps.PROCESSLIST_USER,'@',pps.PROCESSLIST_HOST),
			replace(pps.NAME,'thread/','')) AS user,
			pps.PROCESSLIST_DB AS db, 
			IF(CHAR_LENGTH(pps.PROCESSLIST_INFO) > 64, REPLACE(CONCAT(LEFT(pps.PROCESSLIST_INFO, 30), ' ... ', RIGHT(pps.PROCESSLIST_INFO, 30)), '\n', ' '), REPLACE(pps.PROCESSLIST_INFO, '\n', ' ')) AS current_statement,
			if(isnull(esc.END_EVENT_ID), esc.TIMER_WAIT,NULL) AS statement_latency,
			esc.LOCK_TIME AS lock_latency,
			if(isnull(esc.END_EVENT_ID),esc.TIMER_WAIT,0) AS sort_time
			from (performance_schema.threads pps
			left join performance_schema.events_statements_current esc
			on (pps.THREAD_ID = esc.THREAD_ID))
									left join performance_schema.session_connect_attrs conattr_pid
									on((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID) and (conattr_pid.ATTR_NAME = '_pid'))
			where pps.PROCESSLIST_ID is not null
			and pps.PROCESSLIST_COMMAND <> 'Daemon'
			`
}

func InnoDBLongParams() []string {
	return []string{"innodb_buffer_pool_read_requests%", "innodb_buffer_pool_write_requests%",
		"innodb_buffer_pool_pages_dirty", "innodb_buffer_pool_reads", "innodb_buffer_pool_writes",
		"innodb_os_log_pending_writes", "handler_read_first", "handler_read_key", "handler_read_next",
		"handler_read_prev", "handler_read_rnd", "handler_read_rnd_next", "innodb_data_pending_fsyncs",
		"innodb_os_log_pending_fsyncs"}
}

func SelectLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'SELECT%', count_star, 0)) AS select_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func InsertsLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'INSERTST%', count_star, 0)) AS insert_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func UpdatesLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'UPDATES%', count_star, 0)) AS update_count
			FROM performance_schema.events_statements_summary_by_digest;`
}

func DeletesLongQuery() string {
	return `SELECT SUM(IF(digest_text LIKE 'DELETES%', count_star, 0)) AS delete_count
			FROM performance_schema.events_statements_summary_by_digest;`
}
