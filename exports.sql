--FORMAT
------------------------------------------------------------------
--cmd | thread | pid | state | user | db | time | lock time

--query

------------------------------------------------------------------
--tip: use CTRL+F/CMD+F to easily find what you're looking for
------------------------------------------------------------------

-- 2023-07-21 14:34:12:
 Sleep 1696 1656 82355 root@localhost 0.0000ms 0µs

 

 Sleep 1713 1673 82418 root@localhost 0.0000ms 2000µs

 

 Sleep 1741 1701 root@localhost db 0.0000ms 7000µs

 

 Query 1742 1702 executing root@localhost db 0.6560ms 0µs

 SELECT COUNT(*) AS ongoing_query_count FROM information_schema.pro
cesslist WHERE COMMAND <> 'Sleep' 

 Query 1743 1703 executing root@localhost db 0.8760ms 3000µs

 SELECT pps.PROCESSLIST_COMMAND AS command, pps.THREAD_ID AS thd
_id, pps.PROCESSLIST_ID AS conn_id, conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state, IF( (pps.NAME in ('thread
/sql/one_connection', 'thread/thread_pool/tp_one_connection')), concat(pps.PROCESSLIST_USER, '@', pps.PROCESSLIST_HOST), rep
lace(pps.NAME, 'thread/', '') ) AS user, pps.PROCESSLIST_DB AS db, pps.PROCESSLIST_INFO AS current_statement, IF(isnull(esc.
END_EVENT_ID), esc.TIMER_WAIT, NULL) AS statement_latency, esc.LOCK_TIME AS lock_latency, IF(isnull(esc.END_EVENT_ID), esc.T
IMER_WAIT, 0) AS sort_time FROM performance_schema.threads pps LEFT JOIN performance_schema.events_statements_current esc ON
 (pps.THREAD_ID = esc.THREAD_ID) LEFT JOIN performance_schema.session_connect_attrs conattr_pid ON ( conattr_pid.PROCESSLIST
_ID = pps.PROCESSLIST_ID AND conattr_pid.ATTR_NAME = '_pid' ) WHERE pps.PROCESSLIST_ID IS NOT NULL AND 

