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

func DisplayProcesslist() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	go connectionSanitiser(ctx, cancel)

	var (
		pause   atomic.Value
		export  atomic.Value
		analyse atomic.Value

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

	pause.Store(false)
	export.Store(false)
	analyse.Store(false)

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

	info_text, _ = text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

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

	queries_lc, _ = linechart.New(
		linechart.YAxisAdaptive(),
		linechart.YAxisFormattedValues(linechart.ValueFormatterRoundWithSuffix("")),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	go dynProcesslist(ctx,
		pl_text,
		info_text,
		acts_bc,
		queries_lc,
		search,
		exclude,
		group,
		&pause,
		&export,
		&analyse)

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
									container.SplitPercent(40),
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
		case keyboard.KeyTab:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case '\\':
			time.Sleep(50 * time.Millisecond)
			search.ReadAndClear()
			exclude.ReadAndClear()
			group.ReadAndClear()
		case '/':
			time.Sleep(50 * time.Millisecond)
			group.ReadAndClear()
		case '+':
			Interval += 100 * time.Millisecond
			cancel()
		case '-':
			if Interval > 100*time.Millisecond {
				Interval -= 100 * time.Millisecond
			}
			cancel()
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
			toanalyse := thread.Read()

			if tokill != "" {
				lookup := GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
				queries.Query(Instances[ActiveConns[0]].Driver, fmt.Sprintf(lookup["kill"](), tokill))
			} else if toanalyse != "" {
				CurrentThread = toanalyse
				analyse.Store(true)

				time.Sleep(100 * time.Millisecond)

				for CurrentQuery == "" {
					time.Sleep(100 * time.Millisecond)
				}

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
	pause *atomic.Value,
	export *atomic.Value,
	analyse *atomic.Value) {

	var (
		processlistChannel chan types.Query = make(chan types.Query)
		metricsChannel     chan types.Query = make(chan types.Query)
	)

	/*
		Launch goroutine for each connection
	*/
	for _, conn := range ActiveConns {
		if len(ActiveConns) == 0 {
			continue
		}

		go fetchProcesslist(ctx,
			conn,
			processlistChannel,
			group,
			pause,
			export,
			analyse)

		go fetchMetrics(ctx,
			conn,
			metricsChannel)
	}

	go displayProcesslist(ctx,
		processlistChannel,
		pl_text,
		pause,
		search,
		exclude,
		analyse)

	go displayMetrics(ctx,
		metricsChannel,
		info_text,
		acts_bc,
		queries_lc)

	<-ctx.Done()
}

func fetchProcesslist(ctx context.Context,
	conn string,
	processlistChannel chan<- types.Query,
	group *textinput.TextInput,
	pause *atomic.Value,
	export *atomic.Value,
	analyse *atomic.Value) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)
		istIte bool         = false

		lookup map[string]func() string = GlobalQueryMap[Instances[conn].DBMS]
		query  types.Query
		order  []int = make([]int, 10)

		ftime     int64
		flocktime int64
		fquery    string
	)

	switch Instances[conn].DBMS {
	case types.MYSQL:
		order = []int{0, 1, 2, 3, 4, 5, 6, 8, 9, 7}

	case types.POSTGRES:
		order = nil

	default:
		order = nil
	}

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			if pause.Load().(bool) {
				if export.Load().(bool) {
					export.Store(false)
					io.ExportProcesslist(query.RawData)
				}
				continue
			}
			query, _ = queries.Query(Instances[conn].Driver, lookup["processlist"]())
			utility.ShuffleQuery(query, order)

			for i := range query.RawData {
				ftime, _ = strconv.ParseInt(query.RawData[i][7], 10, 64)
				query.RawData[i][7] = utility.FormatDuration(ftime)
				flocktime, _ = strconv.ParseInt(query.RawData[i][8], 10, 64)
				query.RawData[i][8] = utility.FormatDuration(flocktime)
				fquery = strings.Join(strings.Fields(strings.ReplaceAll(strings.ReplaceAll(query.RawData[i][9], "\t", " "), "\n", " ")), " ")
				query.RawData[i][9] = fquery
			}

			processlistChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func displayProcesslist(ctx context.Context,
	processlistChannel <-chan types.Query,
	widget *text.Text,
	pause *atomic.Value,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	analyse *atomic.Value) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		header []interface{} = []interface{}{"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock-Time", "Query"}
		format string        = "%-10v %-10v %-10v %-5v %-15v %-15v %-20v %-10v %-10v %-99v\n"

		re *regexp.Regexp = regexp.MustCompile(`(\d+\.\d+)(.)s`) /*` Intellisense fix */

		match         []string = make([]string, 0)
		ps            int64
		sens_filters  bool
		include_regex *regexp.Regexp
		exclude_regex *regexp.Regexp

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
		case query = <-processlistChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			if analyse.Load().(bool) {
				for _, row := range text_buffer {
					if row[1] == CurrentThread {
						CurrentQuery = row[9]
					}
				}
			}
			if pause.Load().(bool) {
				continue
			}
			widget.Reset()
			widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

			colorflipper = -1

			for _, row := range text_buffer {
				var t []string = strings.Split(search.Read(), ",")
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

				if (search.Read() != "" && !include_regex.MatchString(strings.Join(row, ""))) ||
					(exclude.Read() != "" && exclude_regex.MatchString(strings.Join(row, ""))) {
					continue
				}

				match = re.FindStringSubmatch(strings.Join(row, " "))
				ps = utility.DeformatDuration(match[0])

				switch {
				case ps >= 5e9 && ps < 1e10:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
					break
				case ps >= 1e10 && ps < 5e10:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
					break
				case ps >= 5e10 && ps < 1e11:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorMaroon))
					break
				case ps >= 1e11:
					color = text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))
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

				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...), color)
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}

func fetchMetrics(ctx context.Context,
	conn string,
	metricsChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string = GlobalQueryMap[Instances[conn].DBMS]
		query  types.Query
		order  []int = make([]int, 10)

		ftime int64
	)

	switch Instances[conn].DBMS {
	case types.MYSQL:
		order = []int{0, 1, 2, 3, 4, 5, 6}

	case types.POSTGRES:
		order = nil

	default:
		order = nil
	}

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			query, _ = queries.Query(Instances[conn].Driver, lookup["metrics"]())
			utility.ShuffleQuery(query, order)

			for i := range query.RawData {
				ftime, _ = strconv.ParseInt(query.RawData[i][0], 10, 64)
				query.RawData[i][0] = utility.Ftime(int(ftime))
				query.RawData[i] = append([]string{conn}, query.RawData[i]...)
			}

			metricsChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func displayMetrics(ctx context.Context,
	metricsChannel <-chan types.Query,
	widget *text.Text,
	bc *barchart.BarChart,
	lc *linechart.LineChart) {

	var ticker *time.Ticker = time.NewTicker(Interval)
	defer ticker.Stop()

	var (
		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		header []interface{} = []interface{}{"Conn", "Uptime", "Queries", "Threads"}
		format string        = "%-10v %-20v %-10v %-10v\n"

		color        text.WriteOption
		colorflipper int

		qpi_history []float64
		qpi_int     float64
	)

	for {
		if len(ActiveConns) == 0 {
			continue
		}

		select {
		case query = <-metricsChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()
			widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

			colorflipper = -1

			var cumulative_qpi float64 = 0
			for _, row := range text_buffer {
				qpi_int, _ = strconv.ParseFloat(row[2], 64)
				cumulative_qpi += qpi_int

				if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				}

				if row[0] == ActiveConns[0] {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGreen))
				}

				colorflipper *= -1

				widget.Write(fmt.Sprintf(format, utility.SliceToInterface(row[:4])...), color)

				if bc != nil {
					if row[0] == ActiveConns[0] {
						var (
							selects int
							inserts int
							updates int
							deletes int
						)

						selects, _ = strconv.Atoi(row[4])
						inserts, _ = strconv.Atoi(row[5])
						updates, _ = strconv.Atoi(row[6])
						deletes, _ = strconv.Atoi(row[7])

						bc.Values([]int{selects, inserts, updates, deletes}[:],
							utility.Max([]int{selects, inserts, updates, deletes}[:])+15)
					}
				}
			}

			if lc != nil {
				qpi_history = append(qpi_history, cumulative_qpi)

				lc.Series("first", qpi_history,
					linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
					linechart.SeriesXLabels(map[int]string{
						0: "0",
					}),
				)
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}
