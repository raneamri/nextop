package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/nextop/dev"
	"github.com/raneamri/nextop/ui"
)

func main() {

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "--version":
			fmt.Print(version())
			return
		case "-inject":
			dev.InjectDBMS()
			return
		}
	}

	ui.InterfaceLoop()

	fmt.Print(")/")
	return
}

func version() string {
	return "v0.0.3"
}
