package ui

import (
	"context"
	"os"
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
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/raneamri/nextop/db"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

func DisplayMenu() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	help_table1, _ := text.New()
	help_table1.Write(
		"P Processlist\nD DB Dashboard\nM Memory Dashboard\nE Error Log\nL Lock Log\nC Configs\n? Help\nESC Previous Page\nQ Quit",
		text.WriteCellOpts(cell.Bold()))

	help_table2, _ := text.New()
	help_table2.Write(
		"CTRL+D Reload page\n-> Cycle to next connection\n<- Cycle to previous connection\n\\ Clear all filters\n/ Clear group filters\n+ Increase refresh rate by 100ms\n- Decrease refresh rate by 100ms",
		text.WriteCellOpts(cell.Bold()),
	)

	help_table3, _ := text.New()
	help_table3.Write(
		"REPO https://github.com/raneamri/nextop\nAUTHOR Imrane AMRI\nLICENSE ...\n",
		text.WriteCellOpts(cell.Bold()),
	)

	cont, err := container.New(
		t,
		container.ID("menu_screen"),
		container.Border(linestyle.Light),
		container.BorderTitle("NEXTOP (ESC to go back)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Modes"),
						container.PlaceWidget(help_table1),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle("Actions"),
								container.PlaceWidget(help_table2),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle("Other"),
								container.PlaceWidget(help_table3),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.Right(
				container.Border(linestyle.Double),
			),
			container.SplitPercent(65),
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
		case 'm', 'M':
			State = types.MEM_DASHBOARD
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

func DisplayConfigs() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		dbms    string
		dsn     string
		name    string
		log_msg string
	)

	errlog, _ := text.New()
	errlog.Write("\n   Awaiting submission.", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	instlog, _ := text.New()
	settings_txt, _ := text.New()

	/*
		Display configurated settings (container-1)
	*/
	settings_headers := []string{"refresh-rate", "errlog-refresh-rate", "default-group"}
	dir, _ := os.Getwd()
	settings_txt.Write("\n   "+dir+"/nextop.conf\n", text.WriteCellOpts(cell.Bold()))
	for _, header := range settings_headers {
		settings_txt.Write("\n   "+header, text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		settings_txt.Write(": " + io.FetchSetting(header))
	}

	/*
		Display configurated instances (container-2)
	*/
	for _, inst := range Instances {
		instlog.Write("\n   dbms", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + utility.Strdbms(inst.DBMS))
		instlog.Write("   dsn", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.DSN)))
		instlog.Write("   conn-name", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.ConnName)))
		if inst.Driver != nil {
			instlog.Write(" ONLINE", text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
		} else {
			instlog.Write(" OFFLINE", text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
		}
	}

	/*
		Display text boxes (container-3)
	*/
	dbmsin, err := textinput.New(
		textinput.Label("DBMS ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <mysql, oracle ...>"),
	)
	dsnin, err := textinput.New(
		textinput.Label("DSN  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <user:pass@tcp(host:port)/name>"),
	)
	namein, err := textinput.New(
		textinput.Label("NAME ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(45),
		textinput.ExclusiveKeyboardOnFocus(),
		textinput.Border(linestyle.Light),
		textinput.BorderColor(cell.Color(cell.ColorAqua)),
		textinput.PlaceHolder(" <recommended>"),
	)

	cont, err := container.New(
		t,
		container.ID("configs_display"),
		container.Border(linestyle.Light),
		container.BorderTitle("CONFIGS (ESC to go back, ENTER to submit)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusGroupsNext(keyboard.KeyArrowDown, 1),
		container.KeyFocusGroupsPrevious(keyboard.KeyArrowUp, 1),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Status"),
						container.PlaceWidget(errlog),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Input"),
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(dbmsin),
							),
							container.Bottom(
								container.SplitHorizontal(
									container.Top(
										container.PlaceWidget(dsnin),
									),
									container.Bottom(
										container.PlaceWidget(namein),
									),
									container.SplitPercent(50),
								),
							),
							container.SplitPercent(33),
						),
					),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Configs"),
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Instances"),
						container.PlaceWidget(instlog),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Settings"),
						container.PlaceWidget(settings_txt),
					),
					container.SplitPercent(60),
				),
			),
			container.SplitPercent(40),
		),
	)
	if err != nil {
		panic(err)
	}

	/*
		Config has its own keyboard subscriber
	*/
	keyninreader := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case keyboard.KeyEnter:
			/*
				Validate data
			*/
			errlog.Reset()

			dbms = dbmsin.ReadAndClear()
			if utility.Fstr(dbms) == "" || utility.Fstr(dbms) != "MYSQL" {
				log_msg = "\n   Error: Unknown DBMS: " + dbms + "\n"
				errlog.Reset()
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				dsn = dsnin.ReadAndClear()
				name = namein.ReadAndClear()
				return
			}

			dsn = dsnin.ReadAndClear()
			if string(dsn) == "" {
				log_msg = "\n   Error: Blank DSN is invalid."
				errlog.Reset()
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				name = namein.ReadAndClear()
				return
			}

			name = namein.ReadAndClear()
			if name == "" {
				log_msg = "\n   Warning: Blank connection name!"
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
				name = "<unnamed>"
			}

			var inst types.Instance
			inst.DBMS = utility.Dbmsstr(dbms)
			inst.DSN = []byte(dsn)
			inst.ConnName = name

			if !db.Ping(inst) {
				errlog.Reset()
				errlog.Write("\n   Authenticating...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Error: Invalid DSN. Connection closed."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				return
			} else {
				errlog.Reset()
				errlog.Write("\n   Authenticating...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Success! Connection established."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
			}

			inst.Driver, err = db.Connect(inst)
			if err == nil {
				ActiveConns = append(ActiveConns, inst.ConnName)
				if len(ActiveConns) == 1 {
					CurrConn = ActiveConns[0]
				}
			}

			/*
				Keep if valid and sync to prevent dupes
			*/
			Instances[inst.ConnName] = inst
			io.SyncConfig(Instances)

			/*
				Update instances display
			*/
			dynInstanceDisplay(ctx, instlog)

		case keyboard.KeyCtrlD:
			io.SyncConfig(Instances)
			for _, inst := range Instances {
				if inst.Driver == nil {
					var err error
					inst.Driver, err = db.Connect(inst)
					if err == nil {
						ActiveConns = append(ActiveConns, inst.ConnName)
					}
				}
			}
			cancel()
		case '?':
			if len(ActiveConns) > 0 {
				State = types.MENU
				cancel()
			} else {
				errlog.Reset()
				log_msg = "\n   Please make sure to have a minimum of one connection online\n   before changing views."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			}
		case keyboard.KeyEsc:
			if len(ActiveConns) > 0 {
				State = Laststate
				cancel()
			} else {
				errlog.Reset()
				log_msg = "\n   Please make sure to have a minimum of one connection online\n   before changing views."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
			}
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyninreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

/*
container-1 (top left): InnoDB Info
container-2 (right): donuts
*/
func DisplayDbDashboard() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	/*
		InnoDB info (container-1)
	*/
	infoheader := []string{"\n\n         Buffer Pool Size\n", "     Buffer Pool Instance\n\n", "                 Redo Log\n",
		"      InnoDB Logfile Size\n", "       Num InnoDB Logfile\n\n", "          Checkpoint Info\n",
		"           Checkpoint Age\n\n", "      Adaptive Hash Index\n", "       Num AHI Partitions"}
	infolabels, _ := text.New()
	for _, header := range infoheader {
		infolabels.Write(header, text.WriteCellOpts(cell.Bold()))
	}

	bfpheader := []string{"\n\n            Read Requests\n", "           Write Requests\n\n", "               Dirty Data\n\n",
		"            Pending Reads\n", "           Pending Writes\n\n", "    OS Log Pending Writes\n\n", "               Disk Reads\n\n",
		"            Pending Fsync\n", "    OS Log Pending Fsyncs\n"}
	bfplabels, _ := text.New()
	for _, header := range bfpheader {
		bfplabels.Write(header, text.WriteCellOpts(cell.Bold()))
	}

	infotext, _ := text.New()
	infotext.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	bfptext, _ := text.New()
	bfptext.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		Donuts (container-2)
	*/
	checkpoint_donut, err := donut.New(
		donut.HolePercent(65),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(24))),
		donut.Label("Checkpoint Age %", cell.FgColor(cell.ColorWhite)),
	)
	pool_donut, err := donut.New(
		donut.HolePercent(65),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(25))),
		donut.Label("Buffer Pool %", cell.FgColor(cell.ColorWhite)),
	)
	ahi_donut, err := donut.New(
		donut.HolePercent(65),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(26))),
		donut.Label("AHI Ratio %", cell.FgColor(cell.ColorWhite)),
	)
	disk_donut, err := donut.New(
		donut.HolePercent(65),
		donut.CellOpts(cell.FgColor(cell.ColorNumber(27))),
		donut.Label("Disk Read %", cell.FgColor(cell.ColorWhite)),
	)

	go dynDbDashboard(ctx, infotext, bfptext, checkpoint_donut, pool_donut, ahi_donut, disk_donut, Interval)

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
								container.PlaceWidget(infotext),
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
								container.PlaceWidget(bfptext),
							),
							container.SplitPercent(60),
						),
					),
					container.SplitPercent(50),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.PlaceWidget(checkpoint_donut),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.PlaceWidget(pool_donut),
							),
							container.SplitPercent(50),
						),
					),
					container.Bottom(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.PlaceWidget(ahi_donut),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.PlaceWidget(disk_donut),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(50),
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
		case 'm', 'M':
			State = types.MEM_DASHBOARD
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
