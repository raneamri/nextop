package utility

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/raneamri/nextop/types"

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

	tout := cmd.Stdout
	cmd.Stdout = os.Stdout
	cmd.Run()
	cmd.Stdout = tout
}

/*
Pushes an item to the top of instance slice
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
		return instances
	}

	return instances
}

func Max(slice []int) int {
	var max int = 0
	for _, num := range slice {
		if num > max {
			max = num
		}
	}
	return max
}

func PopString(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
}

func SliceToInterface(slice []string) []interface{} {
	var intf []interface{} = make([]interface{}, len(slice))

	for i, v := range slice {
		if len(v) > 256 {
			v = v[:256]
		}
		intf[i] = v
	}

	return intf
}

func LogError(e string) {
	const fpath string = "errors/.error.log"
	f, _ := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()

	_, _ = f.WriteString(e + "\n")
}
