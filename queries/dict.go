package queries

func QueryTypeDict() []string {
	return []string{"processlist",
		"uptime",
		"queries",
		"operations",
		"innodb",
		"ahi",
		"bufferpool",
		"user_alloc",
		"global_alloc",
		"spec_alloc",
		"ramdisk_alloc",
		"checkpoint_info",
		"checkpoint_age",
		"err",
		"locks",
	}
}
