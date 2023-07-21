package errors

import (
	"fmt"
	"os"

	"github.com/raneamri/nextop/types"
)

/*
Specific error for incorrect arguments
*/
func ThrowArgError(arguments []string) {
	fmt.Println("Appropriate startup: <dbms> <dsn> <conn-name> optional:<group-name>")
	os.Exit(1)
}

func ThrowKeybindError(duplicate string) {
	fmt.Println("Keybinding error:")
	fmt.Println("    " + duplicate + " is a duplicate keybind")
	os.Exit(1)
}

func CatchConfigReadError(err error, instances map[string]types.Instance) {
	fmt.Println("Config file broken. Read error.")
	panic(err)
}

func CatchConfigWriteError(err error, inst types.Instance) {
	fmt.Println("Config file broken.")
	panic(err)
}
