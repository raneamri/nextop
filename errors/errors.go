package errors

import (
	"fmt"
	"os"

	"github.com/raneamri/gotop/types"
)

/*
Specific error for incorrect arguments
*/
func ThrowArgError(arguments []string) {
	fmt.Println("Unknown argument(s)/flag(s).")
	fmt.Println("Appropriate arguments: <dbms> <dsn> optional:<conn-name>")
	os.Exit(1)
}

/*
Note: not fully implemented
*/
func CatchConfigReadError(err error, instances []types.Instance) {
	fmt.Println("Config file broken. Attempting to heal...")
	/*
		healConfig()
		instances, err = readConfig()
		if err != nil {
			fmt.Println("Failed. Resetting config...")
			resetConfig()
			instances, err = readConfig()
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
