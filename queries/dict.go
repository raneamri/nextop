package queries

/*
Note that these are ordinally matched to their queries
*/
func QueryTypeDict() []string {
	return []string{"processlist",
		"uptime",
		"queries",
		"operations",
		"thread_analysis",
		"kill",
		"innodb",
		"ahi",
		"bufferpool",
		"threadio",
		"user_alloc",
		"global_alloc",
		"spec_alloc",
		"ramdisk_alloc",
		"checkpoint_info",
		"checkpoint_age",
		"err",
		"locks",
		"replication",
		"transactions",
	}
}
