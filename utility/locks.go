package utility

import (
	"time"
)

var (
	LockHistory []string
	LockQueue   map[string]bool
)

func StartQueue() {
	LockQueue = make(map[string]bool)
}

/*
Signals a lock has been put on
*/
func LockOn(lockName string) {
	LockQueue[lockName] = true
	curr := time.Now()
	locklog := lockName + ": on @ " + string(curr.Format("2006-01-02 15:04:05"))
	LockHistory = append(LockHistory, locklog)
}

/*
Signals a lock has been taken off
*/
func LockOff(lockName string) {
	delete(LockQueue, lockName)
	curr := time.Now()
	locklog := lockName + ": off @ " + string(curr.Format("2006-01-02 15:04:05"))
	LockHistory = append(LockHistory, locklog)
}

/*
Returns true if a lock is active, false otherwise
*/
func LockStatus(lockName string) bool {
	return LockQueue[lockName]
}
