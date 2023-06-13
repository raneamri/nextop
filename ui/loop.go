package ui

import (
	"fmt"
	"time"

	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func InterfaceLoop(instances []types.Instance) {
	/*
		Interface parameters
	*/
	var (
		fps      int           = 60
		interval time.Duration = time.Duration(fps/60) * time.Second
	)

	for 1 == 1 {
		utility.ClearTerminal()
		dashboard := InitDashboard(instances)
		fmt.Println(dashboard.String())
		time.Sleep(interval)
	}
}
