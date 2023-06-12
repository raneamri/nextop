package main

import (
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
