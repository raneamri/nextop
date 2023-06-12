package utility

import (
	"fmt"
	"github.com/raneamri/gotop/types"
	"os"
	"os/exec"
	"runtime"

	_ "github.com/go-sql-driver/mysql"
)

/*
Clears stdin
*/
func ClearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

/*
Pushes an item to the top of a slice
Also provides item with index
*/
func PushInstance(instances []types.Instance, pushing types.Instance) []types.Instance {
	pushing.Ndx = len(instances)
	instances = append(instances, pushing)
	return instances
}

/*
Removes an instance from []Instance by value
Re-indexes slice
*/
func PopInstance(instances []types.Instance, popping types.Instance) []types.Instance {
	var rm int = -1
	for i, it := range instances {
		if it.DBMS == popping.DBMS && it.Host == popping.Host && it.Port == popping.Port && it.User == popping.User {
			rm = i
		}
	}

	if rm != -1 {
		/*
			Create slice omitting rm element
		*/
		newinsts := make([]types.Instance, len(instances)-1)
		copy(newinsts[:rm], instances[:rm])
		copy(newinsts[rm:], instances[rm+1:])

		instances = newinsts
	} else {
		fmt.Println("Instance to pop unfound.")
	}

	/*
		Re-index instances
	*/
	for i, it := range instances {
		it.Ndx = i
	}

	return instances
}
