package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

/*
Clears stdin
*/
func clearTerminal() {
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
Formats a string into block capitals and turns spaces into underscores
*/
func fstr(formattable string) string {
	words := strings.Fields(formattable) // Split the string into words
	formattedWords := make([]string, len(words))

	for i, word := range words {
		upperCaseWord := strings.Map(unicode.ToUpper, word)             // Convert word to block capitals
		formattedWords[i] = strings.ReplaceAll(upperCaseWord, " ", "_") // Replace spaces with underscores
	}

	return strings.Join(formattedWords, "_")
}

/*
Formats time as float64 seconds into hours mins secs
*/
func ftime(duration float64) string {
	/*
		float64 -> time.Duration
	*/
	fduration := time.Duration(duration * float64(time.Second))

	/*
		Extract hours, minutes & seconds
	*/
	hours := int(fduration.Hours())
	minutes := int(fduration.Minutes()) % 60
	seconds := int(fduration.Seconds()) % 60

	/*
		Format and concatenate
	*/
	ftime := fmt.Sprintf("%dh %dmin %ds", hours, minutes, seconds)
	return ftime
}

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
