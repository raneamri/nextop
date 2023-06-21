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
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	_ "github.com/go-sql-driver/mysql"
)

/*
Provides data dynamically to displays.go
Note: add filter arg and rake frows
*/
func dynProcesslist(ctx context.Context, pl *text.Text, acc_pl []string, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pldata := db.GetLongQuery(cpool[0], db.ProcesslistLongQuery())

			var frow string
			for _, row := range pldata {
				/*
					Converting data to usable
				*/
				ftime, _ := strconv.ParseInt(row[8], 10, 64)
				row[8] = utility.FpicoToMs(ftime)
				flocktime, _ := strconv.ParseInt(row[9], 10, 64)
				row[9] = utility.FpicoToUs(flocktime)

				frow = fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[8], row[9], row[7])
				acc_pl = append([]string{frow}, acc_pl...)
			}

			/*
				Cap processlist
			*/
			if len(acc_pl) > 500 {
				acc_pl = acc_pl[:500]
			}

			pl_header := fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n",
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

			parameters := []string{"uptime", "queries", "threads_connected"}
			statuses := db.GetStatus(cpool[0], parameters)

			uptime, _ := strconv.Atoi(statuses[0])
			qps := statuses[1]
			thrds := statuses[2]

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

func dynGraphs(ctx context.Context, lc *linechart.LineChart, bc *barchart.BarChart, queries []float64, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			parameters := []string{"queries"}
			variables := db.GetStatus(cpool[0], parameters)
			qps, _ := strconv.ParseFloat(variables[0], 64)
			queries = append(queries, math.Round(qps))

			selects := db.GetLongQuery(cpool[0], db.SelectLongQuery())
			selects_int, _ := strconv.Atoi(selects[0][0])
			inserts := db.GetLongQuery(cpool[0], db.InsertsLongQuery())
			inserts_int, _ := strconv.Atoi(inserts[0][0])
			updates := db.GetLongQuery(cpool[0], db.UpdatesLongQuery())
			updates_int, _ := strconv.Atoi(updates[0][0])
			deletes := db.GetLongQuery(cpool[0], db.DeletesLongQuery())
			deletes_int, _ := strconv.Atoi(deletes[0][0])
			values := []int{selects_int, inserts_int, updates_int, deletes_int}

			if err := lc.Series("first", queries,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			if err := bc.Values(values, utility.Max(values)); err != nil {
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

func dynDbDashboard(ctx context.Context, dbinfo *text.Text, bfpinfo *text.Text, delay time.Duration, cpool []*sql.DB) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			varparameters := []string{"innodb_buffer_pool_size", "datadir", "innodb_log_file_size",
				"innodb_log_files_in_group", "innodb_adaptive_hash_index"}
			variables := db.GetVariable(cpool[0], varparameters)

			bf_pool_int, _ := strconv.Atoi(variables[0])
			bfpool_size := "\n\n" + utility.BytesToMiB(bf_pool_int) + "\n"
			bfpool_inst := "1" + "\n\n" // omitted since not working in SQL 8+ for now
			redolog := variables[1] + "\n"
			intlogfile, _ := strconv.Atoi(variables[2])
			logfile_size := utility.BytesToMiB(intlogfile) + "\n"
			logfilen := variables[3] + "\n\n"
			checkpoint_info := "0MiB" + "\n" // omitted (not sure what it means)
			checkpoint_age := "0" + "\n\n"   // omitted too
			ahi := variables[4] + "\n"
			ahi_nparts := "1" + "\n" // omitted

			dbinfo.Reset()
			dbinfo.Write(bfpool_size, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(bfpool_inst, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(redolog, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(logfile_size, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(logfilen, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(checkpoint_info, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(checkpoint_age, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(ahi, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(ahi_nparts, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

			varparameters = db.InnoDBLongParams()
			variables = db.GetSchemaVariable(cpool[0], varparameters)

			/*
				Note: fix pendings
			*/
			read_reqs_int, _ := strconv.Atoi(variables[0])
			read_reqs := "\n\n" + utility.Fnum(read_reqs_int) + "\n"
			write_reqs_int, _ := strconv.Atoi(variables[1])
			write_reqs := utility.Fnum(write_reqs_int) + "\n\n"
			dirty_data_bytes, _ := strconv.Atoi(variables[2])
			dirty_data := utility.BytesToMiB(dirty_data_bytes) + "\n\n"
			reads_int, _ := strconv.Atoi(variables[3])
			pending_reads := reads_int
			writes_int, _ := strconv.Atoi(variables[4])
			pending_writes := writes_int
			os_log_pending_writes := variables[5] + "\n\n"
			os_read_first, _ := strconv.Atoi(variables[6])
			os_read_key, _ := strconv.Atoi(variables[7])
			os_read_next, _ := strconv.Atoi(variables[8])
			os_read_prev, _ := strconv.Atoi(variables[9])
			os_read_rnd, _ := strconv.Atoi(variables[10])
			os_read_rnd_next, _ := strconv.Atoi(variables[11])
			disk_reads := utility.Fnum(os_read_first+os_read_key+os_read_next+os_read_prev+os_read_rnd+os_read_rnd_next) + "\n\n"
			pending_fsyncs := variables[12] + "\n"
			os_pending_fsyncs := variables[13] + "\n"

			bfpinfo.Reset()
			bfpinfo.Write(read_reqs, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			bfpinfo.Write(write_reqs, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			bfpinfo.Write(dirty_data, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			bfpinfo.Write(fmt.Sprint(pending_reads)+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			bfpinfo.Write(fmt.Sprint(pending_writes)+"\n\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			bfpinfo.Write(os_log_pending_writes, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			bfpinfo.Write(disk_reads, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			bfpinfo.Write(pending_fsyncs, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			bfpinfo.Write(os_pending_fsyncs, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

		case <-ctx.Done():
			return
		}
	}
}
