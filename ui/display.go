package ui

import (
	"database/sql"
	"fmt"

	"github.com/alexeyco/simpletable"
	_ "github.com/go-sql-driver/mysql"
	"github.com/raneamri/gotop/services"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func InitDashboard(instances []types.Instance, cpool []*sql.DB) *simpletable.Table {
	table := simpletable.New()
	/*
		Set headers
	*/
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Text: "NAME"},
			{Text: "UPTIME"},
			{Text: "QPS"},
		},
	}

	/*
		Catch latent driver errors.
	*/
	for i, instance := range instances {
		if cpool[i] == nil {
			fmt.Println("Null driver. Re-connecting...")
			cpool[i] = services.LaunchInstance(instance)
		}
		row := []*simpletable.Cell{
			{Text: instance.Dbname},
			{Text: utility.Ftime(services.GetUptime(cpool[i]))},
			{Text: fmt.Sprint(services.GetQPS(cpool[i]))},
		}
		table.Body.Cells = append(table.Body.Cells, row)
	}

	/*
		Set table alignment
	*/
	table.SetStyle(simpletable.StyleCompactLite)

	return table
}
