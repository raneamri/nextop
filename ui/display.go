package ui

import (
	"fmt"

	"github.com/alexeyco/simpletable"
	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/gotop/services"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func InitDashboard(instances []types.Instance) *simpletable.Table {
	table := simpletable.New()
	/*
		Set headers
	*/
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "INSTANCE"},
			{Text: "UPTIME"},
			{Text: "QPS"},
		},
	}

	/*
		Fill rows with data
		Note: implement for loop to show data for all instances
	*/
	for _, instance := range instances {
		if instance.Driver == nil {
			fmt.Println("null driver!!!!")
			instance.Driver = services.LaunchInstance(instance)
		}
		row := []*simpletable.Cell{
			{Text: instance.Dbname},
			{Text: utility.Ftime(services.GetUptime(instance.Driver))},
			{Text: fmt.Sprint(services.GetQPS(instance.Driver))},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	/*
		Set table alignment
	*/
	table.SetStyle(simpletable.StyleCompactLite)

	return table
}
