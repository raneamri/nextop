package main

import (
	"fmt"
	"os"
)

/*
Specific error for incorrect arguments
*/
func throwArgError(arguments []string) {
	fmt.Println("Unknown argument(s)/flag(s).")
	fmt.Println("Appropriate arguments: <dbms> <username> -w/ws<pass> <(default=3306)port> <(default=127.0.0.1)host> <(default=none)db-name> --s")
	/*
		Flags yet to be implemented
	*/
	fmt.Println("Flags: -w  -> write password to config file as plaintext\n       -ws -> encrypt and write password safely")
	fmt.Println("       --s -> save login to config")
	os.Exit(1)
}

func catchConfigReadError(err error, instances []Instance) {
	fmt.Println("Config file broken. Attempting to heal...")
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
}

/*
Three step config error handler
Step one is attempting to heal the config file
by removing irregularities
Step two is resetting the config file
Step three is throwing the error
*/
func catchConfigWriteError(err error, inst Instance) {
	fmt.Println("Config file broken. Attempting to heal...")
	healConfig()
	err = writeConfig(inst)
	if err != nil {
		fmt.Println("Failed. Resetting config...")
		resetConfig()
		err := writeConfig(inst)
		if err != nil {
			fmt.Println("Fatal error: ")
			panic(err)
		} else {
			fmt.Println("Success! Configurations fully reset & instance written to config")
		}
	} else {
		fmt.Println("Success! Instance written to config.")
	}
}
