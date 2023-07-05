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
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/raneamri/nextop/db"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:
	9 separate queries
	across 9 goroutines
	updating 4 widgets (3 read only)
*/

/*
Format:

	widget-1 (bottom): processlist
	widget-2 (top-left): uptime & qps
	widget-3 (top-mid): barchart showing sel/ins/del/upd
	widget-4 (top-right): query lifeline
*/
func DisplayProcesslist() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	/*
		widget-1
	*/
	pl_text, _ := text.New()
	pl_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	search, err := textinput.New(
		textinput.Label("Search  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("pl-include-suggestion")),
	)
	exclude, err := textinput.New(
		textinput.Label("Exclude ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("pl-exclude-suggestion")),
	)
	group, err := textinput.New(
		textinput.Label("Group   ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Group name"),
	)

	/*
		widget-2
	*/
	info_text, _ := text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		widget-3
	*/
	acts_bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(24),
			cell.ColorNumber(25),
			cell.ColorNumber(26),
			cell.ColorNumber(27),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorWhite,
			cell.ColorWhite,
			cell.ColorWhite,
			cell.ColorWhite,
		}),
		barchart.LabelColors([]cell.Color{
			cell.ColorWhite,
			cell.ColorWhite,
			cell.ColorWhite,
			cell.ColorWhite,
		}),
		barchart.ShowValues(),
		barchart.BarWidth(6),
		barchart.Labels([]string{
			"Sel",
			"Ins",
			"Upd",
			"Del",
		}),
	)

	/*
		widget-4
	*/
	queries_lc, err := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.YAxisFormattedValues(linechart.ValueFormatterRoundWithSuffix("")),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	go dynProcesslist(ctx, pl_text,
		info_text,
		acts_bc,
		queries_lc,
		search,
		exclude,
		group,
		Interval)

	cont, err := container.New(
		t,
		container.ID("processlist"),
		container.Border(linestyle.Light),
		container.BorderTitle("PROCESSLIST (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle("Active (scrollable)"),
								container.PlaceWidget(info_text),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle("Filters"),
								container.SplitHorizontal(
									container.Top(
										container.PlaceWidget(search),
									),
									container.Bottom(
										container.SplitHorizontal(
											container.Top(
												container.PlaceWidget(exclude),
											),
											container.Bottom(
												container.PlaceWidget(group),
											),
											container.SplitPercent(50),
										),
									),
									container.SplitPercent(33),
								),
							),
							container.SplitPercent(30),
						),
					),
					container.Right(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.PlaceWidget(acts_bc),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.BorderTitle("QPS"),
								container.PlaceWidget(queries_lc),
							),
							container.SplitPercent(32),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Processes"),
				container.PlaceWidget(pl_text),
			),
			container.SplitPercent(40),
		),
	)

	if err != nil {
		panic(err)
	}

	var keyreader func(k *terminalapi.Keyboard) = func(k *terminalapi.Keyboard) {
		/*
			Rate limiter
		*/
		elapsed := time.Since(LastInputTime)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}
		LastInputTime = time.Now()

		switch k.Key {
		case keyboard.KeyArrowLeft:
			CurrRotateLeft()
		case keyboard.KeyArrowRight:
			CurrRotateRight()
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyCtrlD:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case '\\':
			time.Sleep(100 * time.Millisecond)
			search.ReadAndClear()
			exclude.ReadAndClear()
			group.ReadAndClear()
		case '/':
			time.Sleep(100 * time.Millisecond)
			group.ReadAndClear()
		case '+':
			Interval += 100 * time.Millisecond
		case '-':
			Interval -= 100 * time.Millisecond
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

func dynProcesslist(ctx context.Context,
	pl_text *text.Text,
	info_text *text.Text,
	acts_bc *barchart.BarChart,
	queries_lc *linechart.LineChart,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	group *textinput.TextInput,
	delay time.Duration) {

	var (
		processlistChannel chan []string  = make(chan []string)
		infoChannel        chan []string  = make(chan []string)
		barchartChannel    chan [4]int    = make(chan [4]int)
		linechartChannel   chan []float64 = make(chan []float64)
	)

	go fetchProcesslist(ctx, processlistChannel, group, delay)
	go writeProcesslist(ctx, pl_text, processlistChannel, search, exclude)

	go fetchProcesslistInfo(ctx, infoChannel, delay)
	go writeProcesslistInfo(ctx, info_text, infoChannel, delay)

	go fetchProcesslistBarchart(ctx, barchartChannel, delay)
	go writeProcesslistBarchart(ctx, acts_bc, barchartChannel, delay)

	go fetchProcesslistLinechart(ctx, linechartChannel, delay)
	go writeProcesslistLinechart(ctx, queries_lc, linechartChannel, delay)

	<-ctx.Done()
}

/*
container-1
*/
func fetchProcesslist(ctx context.Context,
	processlistChannel chan<- []string,
	group *textinput.TextInput,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup map[string]func() string
		pldata [][]string
		/*
			Formatting variables
		*/
		ftime     int64
		flocktime int64
		fquery    string
		/*
			Channel message variable
		*/
		messages []string
	)

	for {
		select {
		case <-ticker.C:
			for _, key := range ActiveConns {
				/*
					Handle group filter
				*/
				if group.Read() != "" {
					if Instances[key].Group != group.Read() {
						continue
					}
				}

				/*
					Fetch query using DBMS and keyword
				*/
				lookup = GlobalQueryMap[Instances[key].DBMS]
				pldata = db.GetLongQuery(Instances[key].Driver, lookup["processlist"]())

				for _, row := range pldata {
					/*
						Formatting
					*/
					ftime, _ = strconv.ParseInt(row[8], 10, 64)
					flocktime, _ = strconv.ParseInt(row[9], 10, 64)
					fquery = strings.ReplaceAll(row[7], "\t", " ")
					fquery = strings.ReplaceAll(row[7], "\n", " ")
					if len(fquery) > 30 {
						fquery = fquery[:30] + "..."
					}

					/*
						Line up items & send to channel
					*/
					messages = append(messages, fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n",
						row[0], row[1], row[2], row[3], row[4], row[5], row[6],
						utility.FpicoToMs(ftime), utility.FpicoToUs(flocktime), fquery))
				}
			}

			processlistChannel <- messages
		case <-ctx.Done():
			return
		}
	}
}

func writeProcesslist(ctx context.Context,
	pl_text *text.Text,
	processlistChannel <-chan []string,
	search *textinput.TextInput,
	exclude *textinput.TextInput) {

	var (
		/*
			Parse variables
		*/
		re    *regexp.Regexp
		match []string
		ms    int
		/*
			Display variables
		*/
		message      []string
		headers      string
		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case message = <-processlistChannel:

			pl_text.Reset()
			headers = fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-65v\n",
				"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock Time", "Query")
			pl_text.Write(headers, text.WriteCellOpts(cell.Bold()))

			colorflipper = -1
			for _, process := range message {
				if !strings.Contains(process, search.Read()) || (strings.Contains(process, exclude.Read()) && exclude.Read() != "") {
					continue
				}

				re = regexp.MustCompile(`(\d+)ms`)
				match = re.FindStringSubmatch(process)
				ms, _ = strconv.Atoi(match[1])

				switch ms {
				case 5_000:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorTeal))
					break
				case 10_000:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
					break
				case 30_000:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
					break
				case 60_000:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorMaroon))
					break
				default:
					if colorflipper > 0 {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
					} else {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
					}
					colorflipper *= -1
					break
				}

				pl_text.Write(process, color)
			}

		case <-ctx.Done():
			return
		}
	}
}

/*
container-2
*/
func fetchProcesslistInfo(ctx context.Context,
	infoChannel chan<- []string,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		parameters []string
		statuses   []string
		qps_int    int
		uptime     int

		messages []string
	)

	for {
		select {
		case <-ticker.C:
			for _, key := range ActiveConns {
				parameters = []string{"uptime", "queries", "threads_connected"}
				statuses = db.GetStatus(Instances[key].Driver, parameters)

				uptime, _ = strconv.Atoi(statuses[0])
				qps_int, _ = strconv.Atoi(fmt.Sprint(statuses[1]))

				messages = append(messages, fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
					key, utility.Ftime(uptime), utility.Fnum(qps_int), statuses[2]))
			}

			infoChannel <- messages
			messages = []string{""}

		case <-ctx.Done():
			return
		}
	}
}

func writeProcesslistInfo(ctx context.Context,
	info_text *text.Text,
	infoChannel <-chan []string,
	delay time.Duration) {

	var (
		message []string
		color   text.WriteOption
	)

	for {
		select {
		case message = <-infoChannel:
			headers := fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
				"Connection", "Uptime", "QPS", "Threads")

			info_text.Reset()
			info_text.Write(headers, text.WriteCellOpts(cell.Bold()))
			for i, item := range message {
				if i%2 == 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				}

				info_text.Write(item, color)
			}
		}
	}
}

/*
container-3
*/
func fetchProcesslistBarchart(ctx context.Context,
	barchartChannel chan<- [4]int,
	delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	var (
		messages [4]int
	)

	for {
		select {
		case <-ticker.C:

			/*
				Format data
			*/
			selects := db.GetLongQuery(Instances[CurrConn].Driver, db.MySQLSelectLongQuery())
			selects_int, _ := strconv.Atoi(selects[0][0])
			inserts := db.GetLongQuery(Instances[CurrConn].Driver, db.MySQLInsertsLongQuery())
			inserts_int, _ := strconv.Atoi(inserts[0][0])
			updates := db.GetLongQuery(Instances[CurrConn].Driver, db.MySQLUpdatesLongQuery())
			updates_int, _ := strconv.Atoi(updates[0][0])
			deletes := db.GetLongQuery(Instances[CurrConn].Driver, db.MySQLDeletesLongQuery())
			deletes_int, _ := strconv.Atoi(deletes[0][0])
			messages = [4]int{selects_int, inserts_int, updates_int, deletes_int}

			barchartChannel <- messages

		case <-ctx.Done():
			return
		}
	}

}

func writeProcesslistBarchart(ctx context.Context,
	acts_bc *barchart.BarChart,
	barchartChannel <-chan [4]int,
	delay time.Duration) {

	var (
		message [4]int
	)

	for {
		select {
		case message = <-barchartChannel:

			acts_bc.Values(message[:], utility.Max(message[:])+500)
		}
	}
}

/*
container-4
*/
func fetchProcesslistLinechart(ctx context.Context,
	linechartChannel chan<- []float64,
	delay time.Duration) {

	var (
		parameters []string
		variables  []string
		qps        float64

		messages []float64
	)

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			parameters = []string{"queries"}
			variables = db.GetStatus(Instances[CurrConn].Driver, parameters)
			qps, _ = strconv.ParseFloat(variables[0], 64)
			messages = append(messages, math.Round(qps))

			linechartChannel <- messages

		case <-ctx.Done():
			return
		}
	}

}

func writeProcesslistLinechart(ctx context.Context,
	queries_lc *linechart.LineChart,
	linechartChannel <-chan []float64,
	delay time.Duration) {

	var (
		message []float64
	)

	for {
		select {
		case message = <-linechartChannel:

			queries_lc.Series("first", message,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			)
		}
	}
}
