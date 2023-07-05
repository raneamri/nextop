package ui

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/raneamri/nextop/db"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"

	_ "github.com/go-sql-driver/mysql"
)

func dynMemoryDashboard(ctx context.Context,
	dballoc1_txt *text.Text,
	dballoc2_txt *text.Text,
	usralloc1_txt *text.Text,
	usralloc2_txt *text.Text,
	dballoc_lc *linechart.LineChart,
	hardwalloc1_txt *text.Text,
	hardwalloc2_txt *text.Text,
	alt []float64,
	delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lookup := GlobalQueryMap[Instances[CurrConn].DBMS]
			total_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["global_alloc"]())
			alloc_by_area := db.GetLongQuery(Instances[CurrConn].Driver, lookup["spec_alloc"]())
			usr_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["user_alloc"]())
			ramndisk_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["ramdisk_alloc"]())

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

func dynErrorLog(ctx context.Context,
	log *text.Text,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	err_ot []float64,
	warn_ot []float64,
	sys_ot []float64,
	freqs *linechart.LineChart,
	delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lookup := GlobalQueryMap[Instances[CurrConn].DBMS]
			error_log := db.GetLongQuery(Instances[CurrConn].Driver, lookup["err"]())

			error_log_headers := "Timestamp           " + "Thd " + " Message\n"

			filterfor := search.Read()
			ommit := exclude.Read()

			log.Reset()
			log.Write(error_log_headers, text.WriteCellOpts(cell.Bold()))

			var (
				flip      int = 1
				syscount  int
				warncount int
				errcount  int
			)

			for _, msg := range error_log {
				color := text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				timestamp := msg[0][:strings.Index(msg[0], ".")] + "  "
				thread := msg[1] + "   "
				prio := msg[2]
				//err_code := msg[3] + " "
				//subsys := msg[4] + " "
				logged := prio + ": " + msg[5] + "\n"

				if prio == "System" {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorNavy))
					syscount++
				} else if prio == "Warning" {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
					warncount++
				} else if prio == "Error" {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
					errcount++
				}

				if !strings.Contains(logged, filterfor) || (strings.Contains(logged, ommit) && ommit != "") {
					continue
				}

				if flip > 0 {
					log.Write(timestamp+thread, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				} else {
					log.Write(timestamp+thread, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				}
				flip *= -1
				log.Write(logged, color)
			}

			err_ot = append(err_ot, float64(errcount))
			warn_ot = append(warn_ot, float64(warncount))
			sys_ot = append(sys_ot, float64(syscount))

			if err := freqs.Series("Errors", err_ot,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed)),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			if err := freqs.Series("Warnings", warn_ot,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			if err := freqs.Series("System", sys_ot,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNavy)),
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

func dynLockLog(ctx context.Context,
	log *text.Text,
	err_ot []float64,
	warn_ot []float64,
	other_ot []float64,
	delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//lookup := GlobalQueryMap[Instances[CurrConn].DBMS]
			//lock_log := db.GetLongQuery(Instances[CurrConn].Driver, lookup["locks"]())

		case <-ctx.Done():
			return
		}
	}
}
