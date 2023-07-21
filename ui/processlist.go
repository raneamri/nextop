package ui

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
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
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:
	6 separate queries
	across 8 goroutines
	updating 4 widgets (3 read only)
*/

/*
Format:

	widget-1 (bottom): processlist
	widget-2 (top-left): uptime & qps
	widget-3 (top-mid): barchart showing sel/ins/del/upd
	widget-4 (top-right): query lifeline
	widget-5 (lower-bottom): kill/thread info
*/
var CurrentThread string

func DisplayProcesslist() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		pause  atomic.Value
		export atomic.Value

		pl_text   *text.Text
		info_text *text.Text

		search  *textinput.TextInput
		exclude *textinput.TextInput
		group   *textinput.TextInput

		thread *textinput.TextInput
		kill   *textinput.TextInput

		acts_bc    *barchart.BarChart
		queries_lc *linechart.LineChart
	)

	/*
		Flags
	*/
	pause.Store(false)
	export.Store(false)

	/*
		widget-1
	*/
	pl_text, _ = text.New()
	pl_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	search, _ = textinput.New(
		textinput.Label("Search  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("pl-include-suggestion")),
	)
	exclude, _ = textinput.New(
		textinput.Label("Exclude ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("pl-exclude-suggestion")),
	)
	group, _ = textinput.New(
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
	info_text, _ = text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		widget-3
	*/
	acts_bc, _ = barchart.New(
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
	queries_lc, _ = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.YAxisFormattedValues(linechart.ValueFormatterRoundWithSuffix("")),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	/*
		widget-5
	*/
	thread, _ = textinput.New(
		textinput.Label("Analyse  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Thread ID"),
	)
	kill, _ = textinput.New(
		textinput.Label("Kill ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Conn ID"),
	)

	go dynProcesslist(ctx,
		pl_text,
		info_text,
		acts_bc,
		queries_lc,
		search,
		exclude,
		group,
		Interval,
		&pause,
		&export)

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
								container.BorderTitle("Active"),
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
								container.BorderTitle("Queries/Interval"),
								container.PlaceWidget(queries_lc),
							),
							container.SplitPercent(32),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Processes"),
						container.PlaceWidget(pl_text),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(thread),
							),
							container.Right(
								container.PlaceWidget(kill),
							),
						),
					),
					container.SplitPercent(85),
				),
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
			RotateConnsLeft()
		case keyboard.KeyArrowRight:
			RotateConnsRight()
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
			if Interval > 100*time.Millisecond {
				Interval -= 100 * time.Millisecond
			}
		case '=':
			if pause.Load().(bool) {
				pause.Store(false)
			} else {
				pause.Store(true)
			}
		case '_':
			if pause.Load().(bool) {
				export.Store(true)
			}
		case keyboard.KeyEnter:
			tokill := kill.ReadAndClear()
			toanalyse := thread.ReadAndClear()

			if tokill != "" {
				lookup := GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
				queries.GetLongQuery(Instances[ActiveConns[0]].Driver, fmt.Sprintf(lookup["kill"](), tokill))
			} else if toanalyse != "" {
				CurrentThread = toanalyse
				State = types.THREAD_ANALYSIS
				cancel()
			}

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
	delay time.Duration,
	pause *atomic.Value,
	export *atomic.Value) {

	var (
		processlistChannel chan []string  = make(chan []string)
		infoChannel        chan []string  = make(chan []string)
		barchartChannel    chan [4]int    = make(chan [4]int)
		linechartChannel   chan []float64 = make(chan []float64)
	)

	go fetchProcesslist(ctx, processlistChannel, group, pause, export, delay)
	go writeProcesslist(ctx, pl_text, processlistChannel, search, exclude)

	go fetchProcesslistInfo(ctx, infoChannel, linechartChannel, delay, pause)
	go writeProcesslistInfo(ctx, info_text, infoChannel)

	go fetchProcesslistBarchart(ctx, barchartChannel, delay, pause)
	go writeProcesslistBarchart(ctx, acts_bc, barchartChannel)

	go writeProcesslistLinechart(ctx, queries_lc, linechartChannel)

	<-ctx.Done()
}

/*
container-1
*/
func fetchProcesslist(ctx context.Context,
	processlistChannel chan<- []string,
	group *textinput.TextInput,
	pause *atomic.Value,
	export *atomic.Value,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup      map[string]func() string
		pldata      [][]string
		group_found bool
		/*
			Formatting variables
		*/
		ftime     int64
		flocktime int64
		fquery    string
		builder   strings.Builder = strings.Builder{}
		/*
			Channel message variable
		*/
		messages []string = make([]string, 0, 2048)
	)

	for {
		select {
		case <-ticker.C:
			group_found = false
			if pause.Load().(bool) {
				if export.Load().(bool) {
					io.ExportProcesslist(messages)
					export.Store(false)
				}
				continue
			}
			messages = []string{}
			for i, key := range ActiveConns {
				/*
					Handle group filter
				*/
				if group.Read() != "" {
					if Instances[key].Group != group.Read() {
						if i == len(ActiveConns)-1 && !group_found {
							messages = nil
							processlistChannel <- messages
						}
						continue
					} else {
						group_found = true
					}
				}

				/*
					Fetch query using DBMS and keyword
				*/
				lookup = GlobalQueryMap[Instances[key].DBMS]
				pldata = queries.GetLongQuery(Instances[key].Driver, lookup["processlist"]())

				for _, row := range pldata {
					/*
						Formatting
					*/
					ftime, _ = strconv.ParseInt(row[8], 10, 64)
					flocktime, _ = strconv.ParseInt(row[9], 10, 64)
					fquery = strings.ReplaceAll(row[7], "\t", " ")
					fquery = strings.ReplaceAll(fquery, "\n", " ")
					/*
						Line up items & send to channel
					*/
					builder.WriteString(fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v ",
						row[0], row[1], row[2], row[3], row[4], row[5], row[6],
						utility.FpicoToMs(ftime), utility.FpicoToUs(flocktime)))
					builder.WriteString(fquery + "\n")

					messages = append(messages, builder.String())
					builder.Reset()
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
		re            *regexp.Regexp = regexp.MustCompile(`(\d+\.\d+)ms`)
		match         []string       = make([]string, 0)
		ms            float64
		sens_filters  bool
		include_regex *regexp.Regexp
		exclude_regex *regexp.Regexp
		/*
			Display variables
		*/
		message      []string = make([]string, 0)
		headers      string
		color        text.WriteOption
		colorflipper int
	)

	if io.FetchSetting("case-sensitive-filters") == "true" {
		sens_filters = true
	} else {
		sens_filters = false
	}

	for {
		select {
		case message = <-processlistChannel:
			pl_text.Reset()
			headers = fmt.Sprintf("%-7v %-5v %-5v %-8v %-25v %-20v %-18v %10v %10v %-15v\n",
				"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock-Time", "Query")

			colorflipper = -1

			pl_text.Write(headers, text.WriteCellOpts(cell.Bold()))
			for _, process := range message {
				t := strings.Split(search.Read(), ",")
				for i, word := range t {
					t[i] = strings.TrimSpace(word)
				}
				var pattern string
				if sens_filters {
					pattern = strings.Join(t, "|")
				} else {
					pattern = "(?i)" + strings.Join(t, "|")
				}
				include_regex = regexp.MustCompile(pattern)

				t = strings.Split(exclude.Read(), ",")
				for i, word := range t {
					t[i] = strings.TrimSpace(word)
				}
				if sens_filters {
					pattern = strings.Join(t, "|")
				} else {
					pattern = "(?i)" + strings.Join(t, "|")
				}
				exclude_regex = regexp.MustCompile(pattern)

				if (search.Read() != "" && !include_regex.MatchString(process)) ||
					(exclude.Read() != "" && exclude_regex.MatchString(process)) {
					continue
				}

				match = re.FindStringSubmatch(process)
				ms, _ = strconv.ParseFloat(match[1], 64)

				switch {
				case ms >= 5 && ms < 10:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
					break
				case ms >= 10 && ms < 50:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
					break
				case ms >= 50 && ms < 100:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorMaroon))
					break
				case ms >= 100:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorBlack))
					break
				default:
					if colorflipper < 0 {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
					} else {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
					}
					break
				}
				colorflipper *= -1

				if len(process) > 256 {
					process = process[:256] + "\n"
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
	linechartChannel chan<- []float64,
	delay time.Duration,
	pause *atomic.Value) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup map[string]func() string

		statuses [][]string = make([][]string, 0)
		qps_int  int
		uptime   int

		messages   []string  = make([]string, 0)
		lc_message []float64 = make([]float64, 0)
	)

	for {
		select {
		case <-ticker.C:
			if pause.Load().(bool) {
				continue
			}
			for _, key := range ActiveConns {
				lookup = GlobalQueryMap[Instances[key].DBMS]
				statuses = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["uptime"]())

				uptime, _ = strconv.Atoi(statuses[1][1])
				qps_int, _ = strconv.Atoi(queries.GetLongQuery(Instances[key].Driver, lookup["queries"]())[0][0])
				if Instances[key].ConnName == Instances[ActiveConns[0]].ConnName {
					lc_message = append(lc_message, float64(qps_int))
					if len(lc_message) > 32 {
						lc_message = lc_message[1:]
					}
					linechartChannel <- lc_message
				}

				messages = append(messages, fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
					key, utility.Ftime(uptime), utility.Fnum(qps_int), statuses[0][1]))
			}

			infoChannel <- messages
			messages = []string{}

		case <-ctx.Done():
			return
		}
	}
}

func writeProcesslistInfo(ctx context.Context,
	info_text *text.Text,
	infoChannel <-chan []string) {

	var (
		message      []string = make([]string, 0)
		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case message = <-infoChannel:
			headers := fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
				"Connection", "Uptime", "Queries", "Threads")

			colorflipper = -1

			info_text.Reset()
			info_text.Write(headers, text.WriteCellOpts(cell.Bold()))
			for _, item := range message {
				key := strings.Split(item, " ")[0]
				if key == ActiveConns[0] {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGreen))
				} else if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				}
				colorflipper *= -1

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
	delay time.Duration,
	pause *atomic.Value) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup map[string]func() string

		operations [][]string
		selects    int
		inserts    int
		updates    int
		deletes    int
		messages   [4]int
	)

	for {
		select {
		case <-ticker.C:
			if pause.Load().(bool) {
				continue
			}
			/*
				Format data
			*/
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			operations = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["operations"]())
			selects, _ = strconv.Atoi(operations[0][0])
			inserts, _ = strconv.Atoi(operations[0][1])
			updates, _ = strconv.Atoi(operations[0][2])
			deletes, _ = strconv.Atoi(operations[0][3])
			messages = [4]int{selects, inserts, updates, deletes}

			barchartChannel <- messages

		case <-ctx.Done():
			return
		}
	}

}

func writeProcesslistBarchart(ctx context.Context,
	acts_bc *barchart.BarChart,
	barchartChannel <-chan [4]int) {

	var (
		message [4]int
	)

	for {
		select {
		case message = <-barchartChannel:

			acts_bc.Values(message[:], utility.Max(message[:])+15)
		}
	}
}

func writeProcesslistLinechart(ctx context.Context,
	queries_lc *linechart.LineChart,
	linechartChannel <-chan []float64) {

	var (
		message []float64 = make([]float64, 0)
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
