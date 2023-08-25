package ui

import (
	"context"
	"fmt"
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
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:
	3 goroutines
	managing 1 query
	and medium duty formatting
*/

/*
Format:
	widget-1 (top-left): filters
	widget-2 (bottom): error log
	widget-3 (top right): linegraphs
*/

func DisplayErrorLog() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	/*
		Error log (widget-1)
	*/
	log, _ := text.New(
		text.WrapAtRunes(),
	)

	/*
		Error-type frequencies linechart (widget-2)
	*/
	frequencies, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	/*
		Filters (widget-3)
	*/
	search, err := textinput.New(
		textinput.Label("Search  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("err-include-suggestion")),
	)
	exclude, err := textinput.New(
		textinput.Label("Exclude ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" Suggested: "+io.FetchSetting("err-exclude-suggestion")),
	)

	log.Write("\n   Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynErrorLog(ctx, log, search, exclude, frequencies, Interval)

	cont, err := container.New(
		t,
		container.ID("err_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("ERROR LOG (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.PlaceWidget(log),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Filters"),
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(search),
							),
							container.Bottom(
								container.PlaceWidget(exclude),
							),
							container.SplitPercent(50),
						),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Statistics"),
						container.PlaceWidget(frequencies),
					),
					container.SplitPercent(40),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Log"),
				container.PlaceWidget(log),
			),
			container.SplitPercent(30),
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
		case keyboard.KeyCtrlD:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case '?':
			State = types.MENU
			cancel()
		case '\\':
			time.Sleep(100 * time.Millisecond)
			search.ReadAndClear()
			exclude.ReadAndClear()
		case '+':
			Interval += 100 * time.Millisecond
		case '-':
			Interval -= 100 * time.Millisecond
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}

func dynErrorLog(ctx context.Context,
	log *text.Text,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	freqs *linechart.LineChart,
	delay time.Duration) {

	var (
		errorChannel chan [][4]string = make(chan [][4]string)
	)

	go fetchErrors(ctx, errorChannel, delay)
	go writeErrors(ctx, log, search, exclude, freqs, errorChannel)

	<-ctx.Done()
}

func fetchErrors(ctx context.Context,
	errorChannel chan<- [][4]string,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup    map[string]func() string = make(map[string]func() string)
		error_log [][]string               = make([][]string, 0)
		/*
			Channel message variable
		*/
		messages [][4]string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			error_log = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["err"]())

			//fmt.Sprintf("%-20v %-5v %-55v\n",

			messages = make([][4]string, len(error_log)+1)

			messages[0][0] = "Timestamp"
			messages[0][1] = "Thd"
			messages[0][2] = "Message"
			messages[0][3] = " "

			for i, msg := range error_log {
				messages[i+1][0] = msg[0][:strings.Index(msg[0], ".")]
				messages[i+1][1] = msg[1]
				messages[i+1][2] = msg[2]
				messages[i+1][3] = msg[5]
			}

			errorChannel <- messages
		case <-ctx.Done():
			return
		}
	}
}

func writeErrors(ctx context.Context,
	log *text.Text,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	freqs *linechart.LineChart,
	errorChannel <-chan [][4]string) {

	var (
		msg          string
		lc_msg       [3]float64
		sens_filters bool
		/*
			Display variables
		*/
		messages    [][4]string = make([][4]string, 0)
		lc_messages [3][]float64

		color        text.WriteOption
		colorflipper int = -1
	)

	if io.FetchSetting("case-sensitive-filters") == "true" {
		sens_filters = true
	} else {
		sens_filters = false
	}

	for {
		select {
		case messages = <-errorChannel:
			log.Reset()
			for i, row := range messages {
				msg = fmt.Sprintf("%-20v %-6v %-8v %-55v\n", row[0], row[1], row[2], row[3])

				if i == 0 {
					log.Write(msg, text.WriteCellOpts(cell.Bold()))
				} else {
					if sens_filters {
						if !strings.Contains(msg, search.Read()) || (strings.Contains(msg, exclude.Read()) && exclude.Read() != "") {
							continue
						}
					} else {
						if !strings.Contains(strings.ToLower(msg), strings.ToLower(search.Read())) ||
							(strings.Contains(strings.ToLower(msg), strings.ToLower(exclude.Read())) && exclude.Read() != "") {
							continue
						}
					}

					if colorflipper > 0 {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
					} else {
						color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
					}

					switch row[2] {
					case "Error":
						color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
						lc_msg[0]++
					case "Warning":
						color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
						lc_msg[1]++
					case "System":
						color = text.WriteCellOpts(cell.FgColor(cell.ColorBlue))
						lc_msg[2]++
					default:
						break
					}

					log.Write(msg, color)
				}
			}

			for i, t := range lc_msg {
				lc_messages[i] = append(lc_messages[i], t)
			}

			if err := freqs.Series("Errors", lc_messages[0],
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed)),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			if err := freqs.Series("Warnings", lc_messages[1],
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			if err := freqs.Series("System", lc_messages[2],
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNavy)),
				linechart.SeriesXLabels(map[int]string{
					0: "0",
				}),
			); err != nil {
				panic(err)
			}

			lc_msg = [3]float64{0, 0, 0}

		case <-ctx.Done():
			return
		}
	}
}
