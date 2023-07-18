package queries

func MySQLFuncDict() []func() string {
	return []func() string{MySQLProcesslist,
		MySQLUptime,
		MySQLQueries,
		MySQLOperationCount,
		MySQLInnoDB,
		MySQLInnoDBAHI,
		MySQLBufferpool,
		MySQLThreadIO,
		MySQLUserMemory,
		MySQLGlobalAllocated,
		MySQLSpecificAllocated,
		MySQLRamNDisk,
		MySQLCheckpointInfo,
		MySQLCheckpointAgePct,
		MySQLErrorLog,
		MySQLLocks,
	}
}

func MapMySQL(MySQLQueries map[string]func() string) {
	types := QueryTypeDict()
	funcs := MySQLFuncDict()

	for i, query := range types {
		MySQLQueries[query] = funcs[i]
	}
}

func MySQLProcesslist() string {
	return `SELECT 
				pps.PROCESSLIST_COMMAND AS command,
				pps.THREAD_ID AS thd_id,
				pps.PROCESSLIST_ID AS conn_id,
				conattr_pid.ATTR_VALUE AS pid,
				pps.PROCESSLIST_STATE AS state,
				IF(
					(pps.NAME in ('thread/sql/one_connection', 'thread/thread_pool/tp_one_connection')),
					concat(pps.PROCESSLIST_USER, '@', pps.PROCESSLIST_HOST),
					replace(pps.NAME, 'thread/', '')
				) AS user,
				pps.PROCESSLIST_DB AS db,
				pps.PROCESSLIST_INFO AS current_statement,
				IF(isnull(esc.END_EVENT_ID), esc.TIMER_WAIT, NULL) AS statement_latency,
				esc.LOCK_TIME AS lock_latency,
				IF(isnull(esc.END_EVENT_ID), esc.TIMER_WAIT, 0) AS sort_time
			FROM 
				performance_schema.threads pps
			LEFT JOIN 
				performance_schema.events_statements_current esc ON (pps.THREAD_ID = esc.THREAD_ID)
			LEFT JOIN 
				performance_schema.session_connect_attrs conattr_pid ON (
					conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID AND 
					conattr_pid.ATTR_NAME = '_pid'
				)
			WHERE 
				pps.PROCESSLIST_ID IS NOT NULL AND 
				pps.PROCESSLIST_COMMAND <> 'Daemon';
			`
}

func MySQLUptime() string {
	return `SHOW STATUS WHERE Variable_name IN ('uptime', 'threads_connected');`
}

func MySQLQueries() string {
	return `SELECT COUNT(*) AS ongoing_query_count
			FROM information_schema.processlist
			WHERE COMMAND <> 'Sleep';`
}

func MySQLOperationCount() string {
	return `SELECT
		    (SELECT COUNT(*) FROM performance_schema.events_statements_current WHERE digest_text LIKE 'SELECT%' AND thread_id IS NOT NULL) AS ongoing_select_count,
			(SELECT COUNT(*) FROM performance_schema.events_statements_current WHERE digest_text LIKE 'INSERT%' AND thread_id IS NOT NULL) AS ongoing_insert_count,
			(SELECT COUNT(*) FROM performance_schema.events_statements_current WHERE digest_text LIKE 'UPDATE%' AND thread_id IS NOT NULL) AS ongoing_update_count,
			(SELECT COUNT(*) FROM performance_schema.events_statements_current WHERE digest_text LIKE 'DELETE%' AND thread_id IS NOT NULL) AS ongoing_delete_count;`
}

func MySQLInnoDB() string {
	return `SELECT
			FORMAT_BYTES((
			SELECT variable_value
			FROM performance_schema.global_variables
			WHERE variable_name = 'innodb_buffer_pool_size'
			)) AS BP_Size,
			(
			SELECT variable_value
			FROM performance_schema.global_variables
			WHERE variable_name = 'innodb_buffer_pool_instances'
			) AS BP_instances,
			CONCAT(
			FORMAT_BYTES(STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'),
			' / ',
			FORMAT_BYTES(
				(SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE VARIABLE_NAME = 'innodb_log_file_size')
				* (SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE VARIABLE_NAME = 'innodb_log_files_in_group')
			)
			) AS CheckpointInfo,
			ROUND(
			(
			(
			(SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"' FROM performance_schema.log_status)
			/ ((SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE VARIABLE_NAME = 'innodb_log_file_size') * (SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE VARIABLE_NAME = 'innodb_log_files_in_group'))
			) * 100
			),
			2
			) AS CheckpointAge,
			FORMAT_BYTES((SELECT VARIABLE_VALUE FROM performance_schema.global_variables WHERE variable_name = 'innodb_log_file_size')) AS InnoDBLogFileSize,
			(
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE variable_name = 'innodb_log_files_in_group'
			) AS NbFiles,
			(
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_status
			WHERE VARIABLE_NAME = 'Innodb_redo_log_enabled'
			) AS RedoEnabled
			FROM performance_schema.log_status;`
}

func MySQLInnoDBAHI() string {
	return `SELECT
			(SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_adaptive_hash_index'
			) AS AHIEnabled,
			(
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_adaptive_hash_index_parts'
			) AS AHIParts,
			ROUND(
			((SELECT VARIABLE_VALUE
			FROM sys.metrics
			WHERE VARIABLE_NAME = 'adaptive_hash_searches'
			) /
			((SELECT VARIABLE_VALUE
			FROM sys.metrics
			WHERE VARIABLE_NAME = 'adaptive_hash_searches_btree'
			) + (
			SELECT VARIABLE_VALUE
			FROM sys.metrics
			WHERE VARIABLE_NAME = 'adaptive_hash_searches'
			))
			) * 100,
			2
			) AS AHIRatio;`
}

func MySQLBufferpool() string {
	return `SELECT
				MAX(CASE WHEN variable_name = 'innodb_buffer_pool_read_requests' THEN variable_value END) AS innodb_buffer_pool_read_requests,
				MAX(CASE WHEN variable_name = 'innodb_buffer_pool_write_requests' THEN variable_value END) AS innodb_buffer_pool_write_requests,
				MAX(CASE WHEN variable_name = 'innodb_buffer_pool_pages_dirty' THEN variable_value END) AS innodb_buffer_pool_pages_dirty,
				MAX(CASE WHEN variable_name = 'innodb_buffer_pool_reads' THEN variable_value END) AS innodb_buffer_pool_reads,
				MAX(CASE WHEN variable_name = 'innodb_buffer_pool_writes' THEN variable_value END) AS innodb_buffer_pool_writes,
				MAX(CASE WHEN variable_name = 'innodb_os_log_pending_writes' THEN variable_value END) AS innodb_os_log_pending_writes,
				MAX(CASE WHEN variable_name = 'handler_read_first' THEN variable_value END) AS handler_read_first,
				MAX(CASE WHEN variable_name = 'handler_read_key' THEN variable_value END) AS handler_read_key,
				MAX(CASE WHEN variable_name = 'handler_read_next' THEN variable_value END) AS handler_read_next,
				MAX(CASE WHEN variable_name = 'handler_read_prev' THEN variable_value END) AS handler_read_prev,
				MAX(CASE WHEN variable_name = 'handler_read_rnd' THEN variable_value END) AS handler_read_rnd,
				MAX(CASE WHEN variable_name = 'handler_read_rnd_next' THEN variable_value END) AS handler_read_rnd_next,
				MAX(CASE WHEN variable_name = 'innodb_data_pending_fsyncs' THEN variable_value END) AS innodb_data_pending_fsyncs,
				MAX(CASE WHEN variable_name = 'innodb_os_log_pending_fsyncs' THEN variable_value END) AS innodb_os_log_pending_fsyncs
			FROM performance_schema.global_status
			WHERE variable_name IN (
				'innodb_buffer_pool_read_requests',
				'innodb_buffer_pool_write_requests',
				'innodb_buffer_pool_pages_dirty',
				'innodb_buffer_pool_reads',
				'innodb_buffer_pool_writes',
				'innodb_os_log_pending_writes',
				'handler_read_first',
				'handler_read_key',
				'handler_read_next',
				'handler_read_prev',
				'handler_read_rnd',
				'handler_read_rnd_next',
				'innodb_data_pending_fsyncs',
				'innodb_os_log_pending_fsyncs'
			);`
}

func MySQLThreadIO() string {
	return ``
}

func MySQLUserMemory() string {
	return `SELECT user, current_allocated, current_max_alloc
			FROM sys.memory_by_user_by_current_bytes
			WHERE user != "background";`
}

func MySQLGlobalAllocated() string {
	return `SELECT total_allocated FROM sys.memory_global_total;`
}

func MySQLSpecificAllocated() string {
	return `SELECT SUBSTRING_INDEX(event_name,'/',2) AS code_area,
			format_bytes(SUM(current_alloc)) AS current_alloc,
			sum(current_alloc) current_alloc_num
			FROM sys.x$memory_global_by_current_bytes
			GROUP BY SUBSTRING_INDEX(event_name,'/',2)
			ORDER BY SUM(current_alloc) DESC;`
}

func MySQLRamNDisk() string {
	return `SELECT event_name,
			format_bytes(CURRENT_NUMBER_OF_BYTES_USED) AS current_alloc,
			format_bytes(HIGH_NUMBER_OF_BYTES_USED) AS high_alloc
			FROM performance_schema.memory_summary_global_by_event_name
			WHERE event_name LIKE 'memory/temptable/%';`
}

func MySQLCheckpointInfo() string {
	return `SELECT CONCAT(
			(
			SELECT FORMAT_BYTES(
			STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
			)
			FROM performance_schema.log_status),
			" / ",
			format_bytes(
			(SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_file_size'
			)  * (
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_files_in_group'))
			) CheckpointInfo;`
}

func MySQLCheckpointAgePct() string {
	return `SELECT ROUND(((
			SELECT STORAGE_ENGINES->>'$."InnoDB"."LSN"' - STORAGE_ENGINES->>'$."InnoDB"."LSN_checkpoint"'
			FROM performance_schema.log_status) / ((
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_file_size'
			) * (
			SELECT VARIABLE_VALUE
			FROM performance_schema.global_variables
			WHERE VARIABLE_NAME = 'innodb_log_files_in_group')) * 100));`
}

func MySQLErrorLog() string {
	return `SELECT *, cast(unix_timestamp(logged)*1000000 as unsigned) logged_int FROM performance_schema.error_log`
}

func MySQLLocks() string {
	return `SELECT
			r.trx_id waiting_trx_id,
			r.trx_mysql_thread_id waiting_thread,
			r.trx_query waiting_query,
			b.trx_id blocking_trx_id,
			b.trx_mysql_thread_id blocking_thread,
			b.trx_query blocking_query
			FROM       performance_schema.data_lock_waits w
			INNER JOIN information_schema.innodb_trx b
			ON b.trx_id = w.blocking_engine_transaction_id
			INNER JOIN information_schema.innodb_trx r
			ON r.trx_id = w.requesting_engine_transaction_id;`
}
