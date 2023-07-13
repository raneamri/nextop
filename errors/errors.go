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

/*
Note: not fully implemented
*/
func CatchConfigReadError(err error, instances map[string]types.Instance) {
	fmt.Println("Config file broken. Attempting to heal...")
	panic(err)
}

/*
Three step config error handler
Step one is attempting to heal the config file
by removing irregularities
Step two is resetting the config file
Step three is throwing the error
Note: not implemented fully
*/
func CatchConfigWriteError(err error, inst types.Instance) {
	fmt.Println("Config file broken. Attempting to heal...")
	/*
		HealConfig()
		err = WriteConfig(inst)
		if err != nil {
			fmt.Println("Failed. Resetting config...")
			ResetConfig()
			err := WriteConfig(inst)
			if err != nil {
				fmt.Println("Fatal error: ")
				panic(err)
			} else {
				fmt.Println("Success! Configurations fully reset & instance written to config")
			}
		} else {
			fmt.Println("Success! Instance written to config.")
		}
	*/
}
