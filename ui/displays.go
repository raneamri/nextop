package ui

import (
	"context"
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
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/raneamri/gotop/db"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Draws the main menu
*/
func DrawMenu(t *tcell.Terminal) {
	/*
		Prepare context to leave state with signal
	*/
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("main_menu"),
		container.Border(linestyle.Light),
		container.BorderTitle("GOTOP (? for help)"),
	)
	if err != nil {
		panic(err)
	}

	/*
		Keyboard reader
	*/
	quitter := func(k *terminalapi.Keyboard) {
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
		case '?':
			State = types.HELP
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	/*
		Display loop
	*/
	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}

/*
Draws the help menu, which can either lead back to the previous view or to a desired view
Static display
*/
func DrawHelp(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	help_table1, _ := text.New()
	help_table1.Write(
		"P Processlist\nD DB Dashboard\nM Memory Dashboard\nE Error Log\nL Lock Log\nC Configs\n? Help\nESC Previous Page\nQ Quit",
		text.WriteCellOpts(cell.Bold()))

	help_table2, _ := text.New()
	help_table2.Write(
		"-> Cycle to next connection\n<- Cycle to previous connection",
		text.WriteCellOpts(cell.Bold()),
	)

	help_table3, _ := text.New()
	help_table3.Write(
		"Other:\n",
		text.WriteCellOpts(cell.Bold()),
	)

	cont, err := container.New(
		t,
		container.ID("help_screen"),
		container.Border(linestyle.Light),
		container.BorderTitle("HELP (ESC to go back)"),
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
	)
	if err != nil {
		panic(err)
	}

	keyreader := func(k *terminalapi.Keyboard) {
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

/*
Format of this display is:

	container-1 (bottom): txt holds parsed processlist data
	container-2 (top-left): txt shows uptime & qps
	container-3 (top-mid): barchart showing sel/ins/del ...
	container-4 (top-right): graph shows lifeline as graph
*/
func DisplayProcesslist(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	/*
		Processlist data (container-1)
	*/
	pl_table, _ := text.New()
	pl_table.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynProcesslist(ctx, pl_table, Interval*2)

	/*
		QPS/Uptime data (container-2)
	*/
	tl_table, _ := text.New()
	tl_table.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynQPSUPT(ctx, tl_table, Interval)

	/*
		SEL/INS ... (container-3)
	*/

	bc, err := barchart.New(
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
	if err != nil {
		panic(err)
	}

	/*
		Queries per hour for the past n hours (container-4)
	*/
	var queries []float64

	lc, err := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.YAxisFormattedValues(linechart.ValueFormatterRoundWithSuffix("")),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)
	if err != nil {
		panic(err)
	}

	go dynPLGraphs(ctx, lc, bc, queries, Interval)

	cont, err := container.New(
		t,
		container.ID("processlist"),
		container.Border(linestyle.Light),
		container.BorderTitle("PROCESSLIST (? for help)"),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.PlaceWidget(tl_table),
					),
					container.Right(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.PlaceWidget(bc),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.BorderTitle("QPS"),
								container.PlaceWidget(lc),
							),
							container.SplitPercent(28),
						),
					),
					container.SplitPercent(30),
				),
			),
			container.Bottom(
				/*
					Processlist
				*/
				container.Border(linestyle.Light),
				container.BorderTitle("Processes"),
				container.PlaceWidget(pl_table),
			),
			container.SplitPercent(40),
		),
	)

	if err != nil {
		panic(err)
	}

	keyreader := func(k *terminalapi.Keyboard) {
		switch k.Key {
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
		case keyboard.KeyArrowLeft:
			CurrRotateLeft()
		case keyboard.KeyArrowRight:
			CurrRotateRight()
		case '?':
			State = types.HELP
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

func DisplayConfigs(t *tcell.Terminal, instances []types.Instance) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		dbms    string
		dsn     string
		name    string
		log_msg string
	)

	errlog, _ := text.New()
	instlog, _ := text.New()

	instlog.Reset()
	for _, inst := range instances {
		instlog.Write("\n   dbms", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + utility.Strdbms(inst.DBMS))
		instlog.Write("   dsn", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.DSN)))
		instlog.Write("   conn-name", text.WriteCellOpts(cell.FgColor(cell.ColorBlue)))
		instlog.Write(": " + string((inst.ConnName)))
	}

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
					),
					container.SplitPercent(70),
				),
			),
			container.SplitPercent(40),
		),
	)
	if err != nil {
		panic(err)
	}

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
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				dsn = dsnin.ReadAndClear()
				name = namein.ReadAndClear()
				return
			}

			dsn = dsnin.ReadAndClear()
			if string(dsn) == "" {
				log_msg = "\n   Error: Blank DSN is invalid."
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
				errlog.Write("\n   Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Error: Invalid DSN. Connection closed."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				return
			} else {
				errlog.Write("\n   Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
				time.Sleep(1 * time.Second)
				errlog.Reset()
				log_msg = "\n   Success! Connection established."
				errlog.Write(log_msg, text.WriteCellOpts(cell.FgColor(cell.ColorGreen)))
			}

			ConnPool[inst.ConnName] = db.Connect(inst)
			ActiveConns = append(ActiveConns, inst.ConnName)
			if len(ActiveConns) == 1 {
				CurrConn = ActiveConns[0]
			}

			instances = append(instances, inst)
			instances = io.SyncConfig(instances)

			dynInstanceDisplay(ctx, instlog, instances, Interval)

		case keyboard.KeyEsc:
			State = Laststate
			cancel()
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
func DisplayDbDashboard(t *tcell.Terminal) {
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
	infotext.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))
	bfptext, _ := text.New()
	bfptext.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

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
			State = types.HELP
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

func DisplayMemory(t *tcell.Terminal) {
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
		Slice to hole allocated memory over time
	*/
	var alt []float64

	go dynMemoryDashboard(ctx, dballoc1_txt, dballoc2_txt, usralloc1_txt, usralloc2_txt, dballoc_lc, hardwalloc1_txt, hardwalloc2_txt, alt, Interval)

	cont, err := container.New(
		t,
		container.ID("memory_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("MEMORY (? for help)"),
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
			State = types.HELP
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

func DisplayErrorLog(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		err_ot   []float64
		warn_ot  []float64
		other_ot []float64
	)
	log, _ := text.New()

	go dynErrorLog(ctx, log, err_ot, warn_ot, other_ot, Interval)

	cont, err := container.New(
		t,
		container.ID("err_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("ERROR LOG (? for help)"),
		container.PlaceWidget(log),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Graph (?)"),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.BorderTitle("Pinboard"),
					),
					container.SplitPercent(60),
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
		case 'l', 'L':
			State = types.LOCK_LOG
			cancel()
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case '?':
			State = types.HELP
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

func DisplayLocks(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	log, _ := text.New()
	active_txt, _ := text.New()

	cont, err := container.New(
		t,
		container.ID("lock_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("LOCKS LOG (? for help)"),
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
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case '?':
			State = types.HELP
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
