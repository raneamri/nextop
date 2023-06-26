package ui

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/raneamri/gotop/db"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	_ "github.com/go-sql-driver/mysql"
)

/*
Provides data dynamically to displays.go
Note: add filter arg and rake frows
*/
func dynProcesslist(ctx context.Context, pl *text.Text, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			/*
				Write header
			*/
			pl_header := fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n",
				"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock Time", "Query")

			var flipper int = 1
			for _, key := range ActiveConns {
				pldata := db.GetLongQuery(ConnPool[key], db.ProcesslistLongQuery())

				var frow string
				var ftable []string
				for _, row := range pldata {
					/*
						Converting data to usable
					*/
					cmd := row[0]
					thread := row[1]
					conn := row[2]
					pid := row[3]
					state := row[4]
					user := row[5]
					db := row[6]
					ftime, _ := strconv.ParseInt(row[8], 10, 64)
					time := utility.FpicoToMs(ftime)
					flocktime, _ := strconv.ParseInt(row[9], 10, 64)
					locktime := utility.FpicoToUs(flocktime)
					fquery := strings.ReplaceAll(row[7], "\t", " ")
					fquery = strings.ReplaceAll(row[7], "\n", " ")
					if len(fquery) > 30 {
						fquery = fquery[:30]
					}

					frow = fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n",
						cmd, thread, conn, pid, state, user, db, time, locktime, fquery)
					ftable = append([]string{frow}, ftable...)
				}

				pl.Reset()
				if err := pl.Write(pl_header, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
					panic(err)
				}
				for _, row := range ftable {
					if flipper > 0 {
						err := pl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
						if err != nil {
							panic(err)
						}
						flipper *= -1
					} else {
						err := pl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
						if err != nil {
							panic(err)
						}
						flipper *= -1
					}
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func dynQPSUPT(ctx context.Context, tl *text.Text, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tl_header := fmt.Sprintf("%-11v %-16v %-7v %-5v\n",
				"Connection", "Uptime", "QPS", "Threads")

			var tabled_data []string

			for _, key := range ActiveConns {
				parameters := []string{"uptime", "queries", "threads_connected"}
				statuses := db.GetStatus(ConnPool[key], parameters)

				uptime, _ := strconv.Atoi(statuses[0])
				qps_int, _ := strconv.Atoi(fmt.Sprint(statuses[1]))
				qps := utility.Fnum(qps_int)
				thrds := statuses[2]

				frow := fmt.Sprintf("%-11v %-16v %-7v %-5v\n", key, utility.Ftime(uptime), qps, fmt.Sprint(thrds))
				tabled_data = append(tabled_data, frow)
			}

			tl.Reset()
			if err := tl.Write(tl_header, text.WriteCellOpts(cell.Bold())); err != nil {
				panic(err)
			}

			for i, row := range tabled_data {
				if i%2 == 0 {
					tl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					tl.Write(row, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func dynPLGraphs(ctx context.Context, lc *linechart.LineChart, bc *barchart.BarChart, queries []float64, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			parameters := []string{"queries"}
			variables := db.GetStatus(ConnPool[CurrConn], parameters)
			qps, _ := strconv.ParseFloat(variables[0], 64)
			queries = append(queries, math.Round(qps))

			/*
				Condense into one query
			*/
			selects := db.GetLongQuery(ConnPool[CurrConn], db.SelectLongQuery())
			selects_int, _ := strconv.Atoi(selects[0][0])
			inserts := db.GetLongQuery(ConnPool[CurrConn], db.InsertsLongQuery())
			inserts_int, _ := strconv.Atoi(inserts[0][0])
			updates := db.GetLongQuery(ConnPool[CurrConn], db.UpdatesLongQuery())
			updates_int, _ := strconv.Atoi(updates[0][0])
			deletes := db.GetLongQuery(ConnPool[CurrConn], db.DeletesLongQuery())
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

func dynDbDashboard(ctx context.Context, dbinfo *text.Text, bfpinfo *text.Text,
	checkpoint_donut *donut.Donut, pool_donut *donut.Donut, ahi_donut *donut.Donut,
	disk_donut *donut.Donut, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			varparameters := []string{"innodb_buffer_pool_size", "innodb_buffer_pool_instances", "innodb_log_file_size",
				"innodb_log_files_in_group", "innodb_adaptive_hash_index", "innodb_adaptive_hash_index_parts"}
			variables := db.GetSchemaVariable(ConnPool[CurrConn], varparameters)

			bf_pool_int, _ := strconv.Atoi(variables[0])
			bfpool_size := "\n\n" + utility.BytesToMiB(bf_pool_int) + "\n"
			bfpool_inst := variables[1] + "\n\n"
			redolog := db.GetSchemaStatus(ConnPool[CurrConn], []string{"innodb_redo_log_enabled"})
			intlogfile, _ := strconv.Atoi(variables[2])
			logfile_size := utility.BytesToMiB(intlogfile) + "\n"
			logfilen := variables[3] + "\n\n"
			checkpoint_info_raw := db.GetLongQuery(ConnPool[CurrConn], db.CheckpointInfoLongQuery())
			checkpoint_info := strings.TrimLeft(checkpoint_info_raw[0][0], " ") + "\n"
			checkpoint_age_raw := db.GetLongQuery(ConnPool[CurrConn], db.CheckpointAgePctLongQuery())
			checkpoint_age := checkpoint_age_raw[0][0] + "%\n\n"
			ahi := variables[4] + "\n"
			ahi_nparts := variables[5] + "\n"

			dbinfo.Reset()
			dbinfo.Write(bfpool_size, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(bfpool_inst, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(redolog[0]+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(logfile_size, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(logfilen, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(checkpoint_info, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(checkpoint_age, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
			dbinfo.Write(ahi, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
			dbinfo.Write(ahi_nparts, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))

			varparameters = db.InnoDBLongParams()
			variables = db.GetSchemaStatus(ConnPool[CurrConn], varparameters)

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

			checkpoint_age_int, _ := strconv.Atoi(checkpoint_age_raw[0][0])
			checkpoint_donut.Percent(checkpoint_age_int)
			pool_donut.Percent(50)
			ahi_donut.Percent(50)
			disk_donut.Percent(50)

		case <-ctx.Done():
			return
		}
	}
}

func dynInstanceDisplay(ctx context.Context, instlog *text.Text, instances []types.Instance, delay time.Duration) {

	instlog.Reset()
	for _, inst := range instances {
		instlog.Write("\n   mysql", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + utility.Strdbms(inst.DBMS))
		instlog.Write("   dsn", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.DSN)))
		instlog.Write("   conn-name", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.ConnName)))
	}
}

func dynMemoryDashboard(ctx context.Context, dballoc1_txt *text.Text, dballoc2_txt *text.Text,
	usralloc1_txt *text.Text, usralloc2_txt *text.Text, dballoc_lc *linechart.LineChart, hardwalloc1_txt *text.Text,
	hardwalloc2_txt *text.Text, alt []float64, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			total_alloc := db.GetLongQuery(ConnPool[CurrConn], db.GlobalAllocatedShortQuery())
			alloc_by_area := db.GetLongQuery(ConnPool[CurrConn], db.SpecificAllocatedLongQuery())
			usr_alloc := db.GetLongQuery(ConnPool[CurrConn], db.UserMemoryShortQuery())
			ramndisk_alloc := db.GetLongQuery(ConnPool[CurrConn], db.RamNDiskLongQuery())

			dballoc1_txt.Reset()
			dballoc2_txt.Reset()
			dballoc1_txt.Write("\n\n   Total allocated\n\n", text.WriteCellOpts(cell.Bold()))
			dballoc2_txt.Write("\n\n"+total_alloc[0][0]+"\n\n", text.WriteCellOpts(cell.Bold()))
			for i, chunk := range alloc_by_area {
				dballoc1_txt.Write("   "+strings.TrimLeft(chunk[0], " ")+"\n", text.WriteCellOpts(cell.Bold()))
				if i%2 == 0 {
					dballoc2_txt.Write(strings.TrimLeft(chunk[1], " ")+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					dballoc2_txt.Write(strings.TrimLeft(chunk[1], " ")+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			usralloc1_txt.Reset()
			usralloc2_txt.Reset()
			usralloc1_txt.Write("\n\n   User\n\n", text.WriteCellOpts(cell.Bold()))
			usralloc2_txt.Write("\n\nCurrent  (Max)\n\n", text.WriteCellOpts(cell.Bold()))
			for j, chunk := range usr_alloc {
				usralloc1_txt.Write("   "+chunk[0]+"\n", text.WriteCellOpts(cell.Bold()))
				chunk[2] = strings.ReplaceAll(chunk[2], " ", "")
				if j%2 == 0 {
					usralloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					usralloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			hardwalloc1_txt.Reset()
			hardwalloc2_txt.Reset()
			hardwalloc1_txt.Write("\n\n\n\n                   Disk\n                    RAM", text.WriteCellOpts(cell.Bold()))
			hardwalloc2_txt.Write("\n\nCurrent  (Max)\n\n", text.WriteCellOpts(cell.Bold()))
			for k, chunk := range ramndisk_alloc {
				chunk[2] = strings.ReplaceAll(chunk[2], " ", "")
				if k%2 == 0 {
					hardwalloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					hardwalloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			parts := strings.SplitN(total_alloc[0][0], " ", 2)
			aps, _ := strconv.ParseFloat(parts[0], 64)
			if aps > 0 {
				alt = append(alt, aps)
			}
			if err := dballoc_lc.Series("first", alt,
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

func dynErrorLog(ctx context.Context, log *text.Text, err_ot []float64, warn_ot []float64,
	other_ot []float64, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

			error_log := db.GetLongQuery(ConnPool[CurrConn], db.ErrorLogShortQuery())

			error_log_headers := "Timestamp                 " + "Thd " + "Message\n"

			log.Reset()
			log.Write(error_log_headers, text.WriteCellOpts(cell.Bold()))

			for i, msg := range error_log {
				color := text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				timestamp := msg[0][:strings.Index(msg[0], ".")] + "  "
				thread := msg[1] + "   "
				prio := msg[2]
				//err_code := msg[3] + " "
				//subsys := msg[4] + " "
				logged := prio + ": " + msg[5] + "\n"

				if prio == "System" {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorNavy))
				} else if prio == "Warning" {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
				}

				if i%2 == 0 {
					log.Write(timestamp+thread, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				} else {
					log.Write(timestamp+thread, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				}
				log.Write(logged, color)
			}

		case <-ctx.Done():
			return
		}
	}
}

func dynLockLog(ctx context.Context, log *text.Text, err_ot []float64, warn_ot []float64,
	other_ot []float64, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

		case <-ctx.Done():
			return
		}
	}
}
