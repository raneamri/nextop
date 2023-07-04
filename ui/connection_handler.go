package ui

/*
Management of connection drivers
*/
/*
Rotates connection pool to the right once
*/
func CurrRotateRight() {
	/*
		Check where in the active connections slice curr is
	*/
	var (
		location int
		length   int
	)

	length = len(ActiveConns)
	for i, conn := range ActiveConns {
		if conn == CurrConn {
			location = i
		}
	}

	/*
		Rotate to first connection if current is last in slice
		else increment by one
	*/
	if location == length-1 {
		CurrConn = ActiveConns[0]
	} else {
		CurrConn = ActiveConns[location+1]
	}

	return
}

/*
Rotates connection pool to the left once
*/
func CurrRotateLeft() {
	/*
		Check where in the active connections slice curr is
	*/
	var (
		location int
		length   int
	)

	length = len(ActiveConns)
	for i, conn := range ActiveConns {
		if conn == CurrConn {
			location = i
		}
	}

	/*
		Rotate to first connection if current is last in slice
		else increment by one
	*/
	if location == 0 {
		CurrConn = ActiveConns[length-1]
	} else {
		CurrConn = ActiveConns[location-1]
	}

	return
}
