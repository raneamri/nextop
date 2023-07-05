package ui

import (
	"context"
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
	"github.com/raneamri/nextop/db"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"

	_ "github.com/go-sql-driver/mysql"
)

func DisplayErrorLog() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		err_ot   []float64
		warn_ot  []float64
		other_ot []float64
	)

	/*
		Error log (container-1)
	*/
	log, _ := text.New(
		text.WrapAtRunes(),
	)

	/*
		Error-type frequencies linechart (container-2)
	*/
	frequencies, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	/*
		Filters (container-3)
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

	log.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		Error log can have heavy fetch / refresh time
		So we display an error log instantly to account for that
	*/
	error_log := db.GetLongQuery(Instances[CurrConn].Driver, db.MySQLErrorLogShortQuery())
	error_log_headers := "Timestamp           " + "Thd " + " Message\n"

	log.Reset()
	log.Write(error_log_headers, text.WriteCellOpts(cell.Bold()))

	for i, msg := range error_log {
		color := text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
		timestamp := msg[0][:strings.Index(msg[0], ".")] + "  "
		thread := msg[1] + "   "
		prio := msg[2]
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

	go dynErrorLog(ctx, log, search, exclude, err_ot, warn_ot, other_ot, frequencies, ErrInterval)

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
