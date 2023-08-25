package queries

func PostgresFuncDict() []func() string {
	return []func() string{PostgresProcesslist,
		PostgresUptime,
		PostgresQueries,
		PostgresOperationCount,
		PostgresThreadAnalysis,
		PostgresKill,
		PostgresInnoDB,
		PostgresInnoDBAHI,
		PostgresBufferpool,
		PostgresThreadIO,
		PostgresUserMemory,
		PostgresGlobalAllocated,
		PostgresSpecificAllocated,
		PostgresRamNDisk,
		PostgresCheckpointInfo,
		PostgresCheckpointAgePct,
		PostgresErrorLog,
		PostgresLocks,
		PostgresReplication,
		PostgresTransactions,
	}
}

func MapPostgres(PostgresQueries map[string]func() string) {
	var types []string = QueryTypeDict()
	var funcs []func() string = PostgresFuncDict()

	for i, query := range types {
		PostgresQueries[query] = funcs[i]
	}
}

func PostgresProcesslist() string {
	return ``
}

func PostgresUptime() string {
	return ``
}

func PostgresQueries() string {
	return ``
}

func PostgresOperationCount() string {
	return ``
}

func PostgresThreadAnalysis() string {
	return `EXPLAIN %s;`
}

func PostgresKill() string {
	return `KILL %s;`
}

func PostgresInnoDB() string {
	return ``
}

func PostgresInnoDBAHI() string {
	return ``
}

func PostgresBufferpool() string {
	return ``
}

func PostgresThreadIO() string {
	return ``
}

func PostgresUserMemory() string {
	return ``
}

func PostgresGlobalAllocated() string {
	return ``
}

func PostgresSpecificAllocated() string {
	return ``
}

func PostgresRamNDisk() string {
	return ``
}

func PostgresCheckpointInfo() string {
	return ``
}

func PostgresCheckpointAgePct() string {
	return ``
}

func PostgresErrorLog() string {
	return ``
}

func PostgresLocks() string {
	return ``
}

func PostgresReplication() string {
	return ``
}

func PostgresTransactions() string {
	return ``
}
