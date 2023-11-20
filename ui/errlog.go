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
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

func DisplayErrorLog() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	log, _ := text.New(
		text.WrapAtRunes(),
	)

	frequencies, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

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

	log.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynErrorLog(ctx, log, search, exclude, frequencies)

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
	freqs *linechart.LineChart) {

	var (
		errorChannel chan types.Query = make(chan types.Query)
	)

	go fetchErrors(ctx, errorChannel)
	go writeErrors(ctx, log, search, exclude, freqs, errorChannel)

	<-ctx.Done()
}

func fetchErrors(ctx context.Context,
	errorChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string = make(map[string]func() string)
		query  types.Query

		order []int = make([]int, 10)
	)

	switch Instances[ActiveConns[0]].DBMS {
	case types.MYSQL:
		order = []int{0, 1, 2, 5}

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

			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["err"]())

			utility.ShuffleQuery(query, order)

			for i, row := range query.RawData {
				query.RawData[i][0] = row[0][:strings.Index(row[0], ".")]
			}

			errorChannel <- query
		case <-ctx.Done():
			return
		}
	}
}

func writeErrors(ctx context.Context,
	widget *text.Text,
	search *textinput.TextInput,
	exclude *textinput.TextInput,
	freqs *linechart.LineChart,
	errorChannel <-chan types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		header []interface{} = []interface{}{"Timestamp", "Thd", "Type", "Message"}
		format string        = "%-20v %-5v %-10v %-100v\n"

		sens_filters bool

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
		case query = <-errorChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()
			widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

			var (
				err_count  float64 = 0
				warn_count float64 = 0
				sys_count  float64 = 0
			)

			for _, row := range text_buffer {
				var (
					conc_row string = strings.Join(row, " ")
				)

				if sens_filters {
					if !strings.Contains(conc_row, search.Read()) || (strings.Contains(conc_row, exclude.Read()) && exclude.Read() != "") {
						continue
					}
				} else {
					if !strings.Contains(strings.ToLower(conc_row), strings.ToLower(search.Read())) ||
						(strings.Contains(strings.ToLower(conc_row), strings.ToLower(exclude.Read())) && exclude.Read() != "") {
						continue
					}
				}

				if colorflipper > 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				}
				colorflipper *= -1

				switch row[2] {
				case "Error":
					color = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
					err_count++
				case "Warning":
					color = text.WriteCellOpts(cell.FgColor(cell.ColorYellow))
					warn_count++
				case "System":
					color = text.WriteCellOpts(cell.FgColor(cell.ColorBlue))
					sys_count++
				default:
					break
				}

				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...), color)
			}

			lc_messages[0] = append(lc_messages[0], err_count)
			lc_messages[1] = append(lc_messages[1], warn_count)
			lc_messages[2] = append(lc_messages[2], sys_count)

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

		case <-ctx.Done():
			return
		}
	}
}
