package ui

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/raneamri/gotop/db"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	_ "github.com/go-sql-driver/mysql"
)

/*
Includes middleman functions that allow for display to dynamically update
Note: add filter arg and rake frows
*/
func dynProcesslist(ctx context.Context, pl *text.Text, acc_pl []string, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_, pldata, _ := db.GetProcesslist(cpool[0])

			var frow string
			for _, row := range pldata {
				/*
					Converting data to usable
				*/
				ftime, _ := strconv.ParseInt(row[8], 10, 64)
				row[8] = utility.FpicoToMs(ftime)
				flocktime, _ := strconv.ParseInt(row[9], 10, 64)
				row[9] = utility.FpicoToUs(flocktime)

				frow = fmt.Sprintf("%-7v %-5v %-5v %-7v %-25v %-20v %-12v %10v %10v %-65v\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[8], row[9], row[7])
				acc_pl = append([]string{frow}, acc_pl...)
			}

			pl_header := fmt.Sprintf("%-7v %-5v %-5v %-7v %-25v %-20v %-12v %10v %10v %-65v\n",
				"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock Time", "Query")

			pl.Reset()
			if err := pl.Write(pl_header, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
				panic(err)
			}

			for i, row := range acc_pl {
				/*
					Flipping digit to alternate row color
				*/
				if i%2 == 0 {
					pl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					pl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func dynQPSUPT(ctx context.Context, tl *text.Text, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tl_header := fmt.Sprintf("%-15v %-20v %-5v\n",
				"Uptime", "QPS", "Threads")

			parameters := []string{"uptime", "queries"}
			variables := db.GetStatus(cpool[0], parameters)
			uptime, _ := strconv.Atoi(variables[0])
			qps := variables[1]
			thrds := db.GetThreads(cpool[0])

			tl.Reset()
			if err := tl.Write(tl_header, text.WriteCellOpts(cell.Bold())); err != nil {
				panic(err)
			}
			frow := fmt.Sprintf("%-15v %-20v %-5v", utility.Ftime(uptime), qps, fmt.Sprint(thrds))
			tl.Write(frow, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

		case <-ctx.Done():
			return
		}
	}
}

func dynQPI(ctx context.Context, lc *linechart.LineChart, queries []float64, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			parameters := []string{"queries"}
			variables := db.GetStatus(cpool[0], parameters)
			qps, _ := strconv.ParseFloat(variables[0], 64)
			queries = append(queries, math.Round(qps))

			if err := lc.Series("first", queries,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}

func dynConfigs(ctx context.Context, logt *text.Text, instt *text.Text, err string, suc string, instances []types.Instance, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logt.Reset()
			if err == "" && suc != "" {
				if err := logt.Write(suc, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorRed))); err != nil {
					panic(err)
				}
			} else if suc == "" && err != "" {
				if err := logt.Write("err", text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorLime))); err != nil {
					panic(err)
				}
			}

			instt.Reset()
			for i, inst := range instances {
				fstr := utility.Strdbms(inst.DBMS) + string(inst.DSN) + inst.Dbname
				if i%2 == 0 {
					if err := instt.Write(fstr, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
						panic(err)
					}
				} else {
					if err := instt.Write(fstr, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorGray))); err != nil {
						panic(err)
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
