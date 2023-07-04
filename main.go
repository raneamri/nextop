package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/nextop/ui"
	"github.com/raneamri/nextop/utility"
)

func main() {

	utility.ClearTerminal()

	/*
		Ensure keybinds are functioning
	*/
	ui.ValidateKeybinds()

	ui.InterfaceLoop()

	fmt.Printf("Bye")
}

func Version() string {
	return "v0.0.1"
}
