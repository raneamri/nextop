package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/nextop/dev"
	"github.com/raneamri/nextop/io"
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

	if _, err := os.Stat(".nextop.conf"); err == nil {
	} else if os.IsNotExist(err) {
		fmt.Println("Configuration file doesn't exist. Healing")
		err := io.HealConfig()
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("Error accessing configuration file: ", err)
	}

	ui.InterfaceLoop()

	fmt.Print(")/")
	return
}

func version() string {
	return "v0.0.3"
}
