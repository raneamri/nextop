package ui

func RotateConnsLeft() {
	if len(ActiveConns) <= 1 {
		return
	}

	ActiveConns = append(ActiveConns[1:], ActiveConns[0])
}

func RotateConnsRight() {
	if len(ActiveConns) <= 1 {
		return
	}

	lastIndex := len(ActiveConns) - 1
	ActiveConns = append([]string{ActiveConns[lastIndex]}, ActiveConns[:lastIndex]...)
}
