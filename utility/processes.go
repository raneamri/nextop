package utility

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/raneamri/gotop/types"

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
	instances = append(instances, pushing)
	return instances
}

/*
Removes an instance from []Instance by DSN value
*/
func PopInstance(instances []types.Instance, popping types.Instance) []types.Instance {
	var rm int = -1
	for i, it := range instances {
		if string(it.DSN) == string(popping.DSN) {
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

	return instances
}
