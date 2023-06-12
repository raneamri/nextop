package main

import (
	"fmt"

	"github.com/alexeyco/simpletable"
)

func InitDashboard(instances []Instance) *simpletable.Table {
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
		row := []*simpletable.Cell{
			{Text: instance.Dbname},
			{Text: Ftime(GetUptime(instance.DB))},
			{Text: fmt.Sprint(GetQPS(instance.DB))},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	/*
		Set table alignment
	*/
	table.SetStyle(simpletable.StyleCompactLite)

	return table
}
