package queries

/*
Note that these are ordinally matched to their queries
and that not all of these translate to each DBMS
as this program was designed around MySQL
*/
func QueryTypeDict() []string {
	return []string{"processlist",
		"metrics",
		"thread_analysis",
		"kill",
		"innodbahi",
		"bufferpool",
		"threadio",
		"malloc",
		"globalalloc",
		"specalloc",
		"ramndisk",
		"err",
		"locks",
		"replication",
		"transactions",
	}
}
