package ui

import (
	"context"
	"fmt"
	"math"
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
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:
	7 goroutines
	sustaining 22 queries
	updating 3 read-only widgets

*/

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

	infoheader := []string{"\n\n   Buffer Pool Size\n",
		"   Buffer Pool Instance\n",
		"   Checkpoint Info\n",
		"   Checkpoint Age\n",
		"   InnoDB Logfile Size\n",
		"   Num InnoDB Logfile\n",
		"   Redo Log\n\n",
		"   Adaptive Hash Indexing\n",
		"   AHI Partitions\n"}
	infolabels, _ := text.New()
	for _, header := range infoheader {
		infolabels.Write(header, text.WriteCellOpts(cell.Bold()))
	}

	bfpheader := []string{"\n\n   Read Requests\n",
		"   Write Requests\n\n",
		"   Dirty Data\n\n",
		"   Pending Reads\n",
		"   Pending Writes\n\n",
		"   OS Log Pending Writes\n\n",
		"   Disk Reads\n\n",
		"   Pending Fsync\n",
		"   OS Log Pending Fsyncs\n"}
	bfplabels, _ := text.New()
	for _, header := range bfpheader {
		bfplabels.Write(header, text.WriteCellOpts(cell.Bold()))
	}

	innodb_text, _ := text.New()
	innodb_text.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	bufferp_text, _ := text.New()
	bufferp_text.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	thdio_text, _ := text.New()
	thdio_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	checkpoint_donut, err := donut.New(
		donut.HolePercent(35),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(24))),
		donut.Label("Checkpoint Age %", cell.FgColor(cell.ColorWhite)),
	)
	pool_donut, err := donut.New(
		donut.HolePercent(35),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(25))),
		donut.Label("Buffer Pool %", cell.FgColor(cell.ColorWhite)),
	)
	ahi_donut, err := donut.New(
		donut.HolePercent(35),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(26))),
		donut.Label("AHI Ratio %", cell.FgColor(cell.ColorWhite)),
	)
	disk_donut, err := donut.New(
		donut.HolePercent(35),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(27))),
		donut.Label("Disk Read %", cell.FgColor(cell.ColorWhite)),
	)

	go dynDbDashboard(ctx, innodb_text, bufferp_text, thdio_text, checkpoint_donut, pool_donut, ahi_donut, disk_donut, Interval)

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
				/*
					container.SplitVertical(
						container.Left(
							container.Border(linestyle.Light),
							container.BorderTitle("Thread I/O"),
							container.PlaceWidget(thdio_text),
						),
						container.Right(
							container.SplitHorizontal(
								container.Top(
									container.SplitHorizontal(
										container.Top(
											container.Border(linestyle.Light),
											container.PlaceWidget(checkpoint_donut),
										),
										container.Bottom(
											container.Border(linestyle.Light),
											container.PlaceWidget(pool_donut),
										),
										container.SplitPercent(50),
									),
								),
								container.Bottom(
									container.SplitHorizontal(
										container.Top(
											container.Border(linestyle.Light),
											container.PlaceWidget(ahi_donut),
										),
										container.Bottom(
											container.Border(linestyle.Light),
											container.PlaceWidget(disk_donut),
										),
										container.SplitPercent(50),
									),
								),
								container.SplitPercent(50),
							),
						),
						container.SplitPercent(75),
					),*/
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
	checkpoint_donut *donut.Donut,
	pool_donut *donut.Donut,
	ahi_donut *donut.Donut,
	disk_donut *donut.Donut,
	delay time.Duration) {

	var (
		innodbChannel     chan string   = make(chan string)
		bufferpoolChannel chan string   = make(chan string)
		donutChannel      chan [4]int   = make(chan [4]int)
		thdioChannel      chan []string = make(chan []string)
	)

	go fetchInnoDb(ctx, innodbChannel, donutChannel, delay)
	go writeInnoDb(ctx, innodb_text, innodbChannel)

	go fetchInnoDbBufferPool(ctx, bufferpoolChannel, delay)
	go writeInnoDbBufferPool(ctx, bufferp_text, bufferpoolChannel)

	go fetchThreadIO(ctx, thdioChannel, delay)
	go writeThreadIO(ctx, thdio_text, thdioChannel)

	go writeInnoDbDonuts(ctx, checkpoint_donut, pool_donut, ahi_donut, disk_donut, donutChannel)

	<-ctx.Done()
}

func fetchInnoDb(ctx context.Context,
	innodbChannel chan<- string,
	donutChannel chan<- [4]int,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup        map[string]func() string
		variables     []string = make([]string, 0)
		ahi_variables []string = make([]string, 0)

		/*
			Channel message variable
		*/
		message       string
		donut_message float64
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			variables = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["innodb"]())[0]
			ahi_variables = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["ahi"]())[0]

			/*
				Compose message
			*/
			variables[2] = strings.TrimLeft(variables[2], " ")
			message = "\n\n" + strings.Join(variables, "\n") + "\n\n" + strings.Join(ahi_variables[:2], "\n")
			donut_message, _ = strconv.ParseFloat(ahi_variables[2], 64)

			innodbChannel <- message
			donutChannel <- [4]int{0, 0, int(math.Round(donut_message)), 0}
			message = ""
		case <-ctx.Done():
			return
		}
	}
}

func writeInnoDb(ctx context.Context,
	innodb_text *text.Text,
	innodbChannel <-chan string) {

	var (
		/*
			Display variables
		*/
		message string
	)

	for {
		select {
		case message = <-innodbChannel:
			innodb_text.Reset()
			innodb_text.Write(message)

		case <-ctx.Done():
			return
		}
	}
}

func fetchInnoDbBufferPool(ctx context.Context,
	bufferpoolChannel chan<- string,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup    map[string]func() string
		variables []string = make([]string, 0)
		/*
			Formatting variables
		*/
		read_reqs_int    int
		write_reqs_int   int
		dirty_data_bytes int
		os_read_first    int
		os_read_key      int
		os_read_next     int
		os_read_prev     int
		os_read_rnd      int
		os_read_rnd_next int
		/*
			Channel message variable
		*/
		message string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			variables = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["bufferpool"]())[0]

			read_reqs_int, _ = strconv.Atoi(variables[0])
			write_reqs_int, _ = strconv.Atoi(variables[1])
			dirty_data_bytes, _ = strconv.Atoi(variables[2])
			os_read_first, _ = strconv.Atoi(variables[6])
			os_read_key, _ = strconv.Atoi(variables[7])
			os_read_next, _ = strconv.Atoi(variables[8])
			os_read_prev, _ = strconv.Atoi(variables[9])
			os_read_rnd, _ = strconv.Atoi(variables[10])
			os_read_rnd_next, _ = strconv.Atoi(variables[11])

			/*
				Note: fix pendings
			*/
			message += "\n\n" + utility.Fnum(read_reqs_int) + "\n"
			message += utility.Fnum(write_reqs_int) + "\n\n"
			message += utility.BytesToMiB(dirty_data_bytes) + "\n\n"
			message += variables[3] + "\n"
			message += variables[4] + "\n\n"
			message += variables[5] + "\n\n"
			message += utility.Fnum(os_read_first+os_read_key+os_read_next+os_read_prev+os_read_rnd+os_read_rnd_next) + "\n\n"
			message += variables[12] + "\n"
			message += variables[13] + "\n"

			bufferpoolChannel <- message
			message = ""
		case <-ctx.Done():
			return
		}
	}
}

func writeInnoDbBufferPool(ctx context.Context,
	bufferp_text *text.Text,
	bufferpoolChannel <-chan string) {

	var (
		/*
			Display variables
		*/
		message string
	)

	for {
		select {
		case message = <-bufferpoolChannel:
			bufferp_text.Reset()
			bufferp_text.Write(message)

		case <-ctx.Done():
			return
		}
	}
}

func fetchThreadIO(ctx context.Context,
	thdioChannel chan<- []string,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup    map[string]func() string
		variables string
		pattern   *regexp.Regexp = regexp.MustCompile(`I/O thread (\d+) state: ([^(]+) \(([^)]*)\)`)

		matches [][]string

		/*
			Channel message variable
		*/
		message []string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			variables = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["threadio"]())[0][2]

			matches = pattern.FindAllStringSubmatch(variables, -1)

			for _, match := range matches {
				parts := strings.SplitN(match[0], ":", 2)
				message = append(message, fmt.Sprintf("\n  %-15v:\n  %-22v\n", parts[0], parts[1]))
			}

			thdioChannel <- message
			message = []string{}

		case <-ctx.Done():
			return
		}
	}
}

func writeThreadIO(ctx context.Context,
	thdio_text *text.Text,
	thdioChannel <-chan []string) {

	var (
		/*
			Display variables
		*/
		message []string

		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case message = <-thdioChannel:
			thdio_text.Reset()

			colorflipper = 1
			for _, line := range message {
				if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				}
				colorflipper *= -1
				thdio_text.Write(line, color)
			}

		case <-ctx.Done():
			return
		}
	}
}

func writeInnoDbDonuts(ctx context.Context,
	checkpoint_donut *donut.Donut,
	pool_donut *donut.Donut,
	ahi_donut *donut.Donut,
	disk_donut *donut.Donut,
	donutChannel <-chan [4]int) {

	var (
		/*
			Display variables
		*/
		message [4]int
	)

	for {
		select {
		case message = <-donutChannel:

			if message[0] != 0 {
				checkpoint_donut.Percent(message[0])
			}
			if message[1] != 0 {
				pool_donut.Percent(message[1])
			}
			if message[2] != 0 {
				ahi_donut.Percent(message[2])
			}
			if message[3] != 0 {
				disk_donut.Percent(message[3])
			}

		case <-ctx.Done():
			return
		}
	}
}
