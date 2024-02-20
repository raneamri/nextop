package ui

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Format:

	widget-1 (top-left): general info
	widget-2 (bottom-left): buffer pool
	widget-3 (center): thread i/o
	widget-4 (right): donuts
*/
func DisplayInnoDbDashboard() {
	t, err := tcell.New()
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go connectionSanitiser(ctx, cancel)

	var infoheader []interface{} = []interface{}{"Buffer Pool Size",
		"Buffer Pool Instance",
		"Checkpoint Info",
		"Checkpoint Age",
		"InnoDB Logfile Size",
		"No. InnoDB Logfile",
		"Redo Log",
		"Adapative Hash Indexing",
		"AHI Partitions"}
	var infoformat string = "\n\n%-10v\n%-10v\n\n%-10v\n%-15v\n\n%-15v\n%-25v\n%-10v\n\n%-10v\n%-10v"
	infolabels, _ := text.New()
	infolabels.Write(fmt.Sprintf(infoformat, infoheader...), text.WriteCellOpts(cell.Bold()))

	var bfpheader []interface{} = []interface{}{"Read Requests",
		"Write Requests",
		"Dirty Data",
		"Pending Reads",
		"Pending Writes",
		"OS Log Pending Writes",
		"Disk Reads",
		"Pending Fsync",
		"OS Log Pending Fsync"}
	var bfpformat string = "\n\n%-10v\n%-10v\n%-10v\n%-15v\n%-15v\n\n%-25v\n%-10v\n\n%-10v\n%-10v"
	bfplabels, _ := text.New()
	bfplabels.Write(fmt.Sprintf(bfpformat, bfpheader...), text.WriteCellOpts(cell.Bold()))

	innodb_text, _ := text.New()
	innodb_text.Write("\n\n Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	bufferp_text, _ := text.New()
	bufferp_text.Write("\n\n Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	thdio_text, _ := text.New()
	thdio_text.Write("\n Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynDbDashboard(ctx, innodb_text, bufferp_text, thdio_text, Interval)

	cont, err := container.New(
		t,
		container.ID("db_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("INNODB DASHBOARD (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Info"),
						container.SplitVertical(
							container.Left(
								container.SplitVertical(
									container.Left(),
									container.Right(
										container.PlaceWidget(infolabels),
									),
									container.SplitPercent(30),
								),
							),
							container.Right(
								container.PlaceWidget(innodb_text),
							),
							container.SplitPercent(60),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Buffer Pool"),
						container.SplitVertical(
							container.Left(
								container.SplitVertical(
									container.Left(),
									container.Right(
										container.PlaceWidget(bfplabels),
									),
									container.SplitPercent(30),
								),
							),
							container.Right(
								container.PlaceWidget(bufferp_text),
							),
							container.SplitPercent(60),
						),
					),
					container.SplitPercent(45),
				),
			),
			container.Right(
				container.Border(linestyle.Light),
				container.BorderTitle("Thread I/O"),
				container.PlaceWidget(thdio_text),
			),
			container.SplitPercent(45),
		),
	)
	if err != nil {
		panic(err)
	}

	keyreader := func(k *terminalapi.Keyboard) {
		elapsed := time.Since(LastInputTime)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}
		LastInputTime = time.Now()

		switch k.Key {
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyTab:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}

func dynDbDashboard(ctx context.Context,
	innodb_text *text.Text,
	bufferp_text *text.Text,
	thdio_text *text.Text,
	delay time.Duration) {

	var (
		innodbChannel     chan types.Query = make(chan types.Query)
		bufferpoolChannel chan types.Query = make(chan types.Query)
		thdioChannel      chan types.Query = make(chan types.Query)
	)

	go fetchInnoDb(ctx,
		innodbChannel)

	go fetchBufferPool(ctx,
		bufferpoolChannel)

	go fetchThreadIO(ctx,
		thdioChannel)

	go displayInnoDb(ctx,
		innodbChannel,
		innodb_text)

	go displayBufferPool(ctx,
		bufferpoolChannel,
		bufferp_text)

	go displayThreadIO(ctx,
		thdioChannel,
		thdio_text)

	<-ctx.Done()
}

func fetchInnoDb(ctx context.Context,
	innodbChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string
		query  types.Query
		order  []int = make([]int, 10)
	)

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			switch Instances[ActiveConns[0]].DBMS {
			case types.MYSQL:
				order = []int{0, 1, 2, 3, 4, 5, 6, 7, 8}

			case types.POSTGRES:
				order = nil

			default:
				order = nil
			}

			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["innodbahi"]())

			utility.ShuffleQuery(query, order)

			for i := range query.RawData {
				query.RawData[i][2] = strings.TrimLeft(query.RawData[i][2], " ")
			}

			innodbChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func displayInnoDb(ctx context.Context,
	innodbChannel <-chan types.Query,
	widget *text.Text) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		format string = "\n\n%-10v\n%-10v\n\n%-20v\n%-25v\n\n%-15v\n%-25v\n%-10v\n\n%-10v\n%-10v"
	)

	for {
		select {
		case query = <-innodbChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()

			for _, row := range text_buffer {
				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...))
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}

func fetchBufferPool(ctx context.Context,
	bufferpoolChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string
		query  types.Query
		order  []int = make([]int, 10)

		read_reqs_int    int
		write_reqs_int   int
		dirty_data_bytes int
		pending_reads    int
		pending_writes   int
		os_read_first    int
		os_read_key      int
		os_read_next     int
		os_read_prev     int
		os_read_rnd      int
		os_read_rnd_next int
	)

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			switch Instances[ActiveConns[0]].DBMS {
			case types.MYSQL:
				order = []int{0, 1, 2, 3, 4, 5, 12, 13, 6, 7, 8, 9, 10, 11}

			case types.POSTGRES:
				order = nil

			default:
				order = nil
			}

			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["bufferpool"]())
			utility.ShuffleQuery(query, order)

			for i := range query.RawData {
				read_reqs_int, _ = strconv.Atoi(query.RawData[i][0])
				write_reqs_int, _ = strconv.Atoi(query.RawData[i][1])
				dirty_data_bytes, _ = strconv.Atoi(query.RawData[i][2])
				pending_reads, _ = strconv.Atoi(query.RawData[i][3])
				pending_writes, _ = strconv.Atoi(query.RawData[i][4])
				os_read_first, _ = strconv.Atoi(query.RawData[i][8])
				os_read_key, _ = strconv.Atoi(query.RawData[i][9])
				os_read_next, _ = strconv.Atoi(query.RawData[i][10])
				os_read_prev, _ = strconv.Atoi(query.RawData[i][11])
				os_read_rnd, _ = strconv.Atoi(query.RawData[i][12])
				os_read_rnd_next, _ = strconv.Atoi(query.RawData[i][13])

				query.RawData[i][0] = utility.Fnum(read_reqs_int)
				query.RawData[i][1] = utility.Fnum(write_reqs_int)
				query.RawData[i][2] = utility.BytesToMiB(dirty_data_bytes)
				query.RawData[i][3] = utility.Fnum(pending_reads)
				query.RawData[i][4] = utility.Fnum(pending_writes)
				query.RawData[i][8] = utility.Fnum(os_read_first + os_read_key + os_read_next + os_read_prev + os_read_rnd + os_read_rnd_next)
				query.RawData[i] = query.RawData[i][:9]
			}

			bufferpoolChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func displayBufferPool(ctx context.Context,
	bufferpoolChannel <-chan types.Query,
	widget *text.Text) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		format string = "\n\n%-10v\n%-10v\n%-20v\n%-25v\n%-15v\n\n%-25v\n%-10v\n\n%-10v\n%-10v"
	)

	for {
		select {
		case query = <-bufferpoolChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()

			for _, row := range text_buffer {
				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...))
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}

func fetchThreadIO(ctx context.Context,
	thdioChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup  map[string]func() string
		query   types.Query
		matches [][]string = make([][]string, 0)

		pattern *regexp.Regexp = regexp.MustCompile(`I/O thread (\d+) state: ([^(]+) \(([^)]*)\)`)
	)

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			switch Instances[ActiveConns[0]].DBMS {
			case types.MYSQL:
				break

			case types.POSTGRES:
				continue

			default:
				continue
			}

			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["threadio"]())

			matches = pattern.FindAllStringSubmatch(query.RawData[0][2], -1)
			query.RawData = matches

			thdioChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func displayThreadIO(ctx context.Context,
	thdioChannel <-chan types.Query,
	widget *text.Text) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		format string = "\n   %-0v \n   %-3v %-31v %-12v\n"

		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case query = <-thdioChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()

			colorflipper = 1
			for _, row := range text_buffer {
				colorflipper *= -1
				if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				}
				widget.Write(fmt.Sprintf(format, utility.SliceToInterface(row)...), color)
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}

}
