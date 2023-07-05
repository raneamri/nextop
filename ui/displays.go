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

func DisplayMemory() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	dballoc1_txt, _ := text.New()
	dballoc2_txt, _ := text.New()
	usralloc1_txt, _ := text.New()
	usralloc2_txt, _ := text.New()
	dballoc_lc, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)
	hardwalloc1_txt, _ := text.New()
	hardwalloc2_txt, _ := text.New()

	/*
		Slice to hold allocated memory over time
	*/
	var alt []float64

	go dynMemoryDashboard(ctx, dballoc1_txt, dballoc2_txt, usralloc1_txt, usralloc2_txt, dballoc_lc, hardwalloc1_txt, hardwalloc2_txt, alt, Interval)

	cont, err := container.New(
		t,
		container.ID("memory_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("MEMORY (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Db Memory Allocation"),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(dballoc1_txt),
							),
							container.Right(
								container.PlaceWidget(dballoc2_txt),
							),
							container.SplitPercent(60),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Users Memory Allocation"),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(usralloc1_txt),
							),
							container.Right(
								container.PlaceWidget(usralloc2_txt),
							),
							container.SplitPercent(60),
						),
					),
					container.SplitPercent(65),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Total Allocated Memory"),
						container.PlaceWidget(dballoc_lc),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Disk / RAM"),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(hardwalloc1_txt),
							),
							container.Right(
								container.PlaceWidget(hardwalloc2_txt),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(65),
				),
			),
			container.SplitPercent(50),
		),
	)
	if err != nil {
		panic(err)
	}

	keyreader := func(k *terminalapi.Keyboard) {
		// Calculate the time elapsed since the last input
		elapsed := time.Since(LastInputTime)

		// Set a minimum cooldown period (e.g., 500 milliseconds)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))

		// If the elapsed time is less than the cooldown period, ignore the input
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}

		// Update the last input time to the current time
		LastInputTime = time.Now()

		switch k.Key {
		case 'p', 'P':
			State = types.PROCESSLIST
			cancel()
		case 'd', 'D':
			State = types.DB_DASHBOARD
			cancel()
		case 'e', 'E':
			State = types.ERR_LOG
			cancel()
		case 'l', 'L':
			State = types.LOCK_LOG
			cancel()
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyCtrlD:
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
		// Calculate the time elapsed since the last input
		elapsed := time.Since(LastInputTime)

		// Set a minimum cooldown period (e.g., 500 milliseconds)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))

		// If the elapsed time is less than the cooldown period, ignore the input
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}

		// Update the last input time to the current time
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

func DisplayLocks() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	log, _ := text.New()
	active_txt, _ := text.New()

	cont, err := container.New(
		t,
		container.ID("lock_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("LOCKS LOG (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Active (?)"),
				container.PlaceWidget(log),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Locks"),
				container.PlaceWidget(active_txt),
			),
			container.SplitPercent(20),
		),
	)
	if err != nil {
		panic(err)
	}

	keyreader := func(k *terminalapi.Keyboard) {
		// Calculate the time elapsed since the last input
		elapsed := time.Since(LastInputTime)

		// Set a minimum cooldown period (e.g., 500 milliseconds)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))

		// If the elapsed time is less than the cooldown period, ignore the input
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}

		// Update the last input time to the current time
		LastInputTime = time.Now()

		switch k.Key {
		case keyboard.KeyCtrlD:
			cancel()
		case 'p', 'P':
			State = types.PROCESSLIST
			cancel()
		case 'd', 'D':
			State = types.DB_DASHBOARD
			cancel()
		case 'm', 'M':
			State = types.MEM_DASHBOARD
			cancel()
		case 'e', 'E':
			State = types.ERR_LOG
			cancel()
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case '?':
			State = types.MENU
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
