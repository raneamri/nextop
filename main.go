package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/nextop/ui"
)

func main() {

	ui.InterfaceLoop()

	fmt.Printf("Bye")
}

func Version() string {
	return "v0.0.2"
}
