package ui

import (
	"context"
	"database/sql"
	"strconv"
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
		"Actions:\n",
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
func DisplayProcesslist(t *tcell.Terminal, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	/*
		Processlist data (container-1)
	*/
	pl_table, _ := text.New()

	/*
		Accumulated process list
	*/
	var acc_pl []string

	go dynProcesslist(ctx, pl_table, acc_pl, Interval, cpool)

	/*
		QPS/Uptime data (container-2)
	*/
	tl_table, _ := text.New()

	go dynQPSUPT(ctx, tl_table, Interval, cpool)

	/*
		SEL/INS ... (container-3)
	*/

	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(25),
			cell.ColorNumber(30),
			cell.ColorNumber(35),
			cell.ColorNumber(40),
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

	stmt := []string{"Queries"}
	minl := db.GetStatus(cpool[0], stmt)
	min, _ := strconv.ParseFloat(minl[0], 64)

	lc, err := linechart.New(
		linechart.YAxisCustomScale(min, min*2),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)
	if err != nil {
		panic(err)
	}

	go dynGraphs(ctx, lc, bc, queries, Interval, cpool)

	cont, err := container.New(
		t,
		container.ID("processlist"),
		container.Border(linestyle.Light),
		container.BorderTitle("PROCESSLIST (? for help, ESC to go back, ,<- -> to navigate)"),
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
								container.BorderTitle("Queries/h"),
								container.PlaceWidget(lc),
							),
							container.SplitPercent(30),
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

func DisplayConfigs(t *tcell.Terminal, instances []types.Instance, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	var (
		dbms  string
		dsn   string
		name  string
		inerr string
		insuc string
	)

	errlog, _ := text.New()
	errlog.Write(insuc, text.WriteCellOpts(cell.FgColor(cell.ColorRed)))

	confdisp, _ := text.New()

	go dynConfigs(ctx, errlog, confdisp, inerr, insuc, instances, Interval)

	dbmsin, err := textinput.New(
		textinput.Label("DBMS ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(35),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	dsnin, err := textinput.New(
		textinput.Label("DSN  ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(35),
		textinput.ExclusiveKeyboardOnFocus(),
	)
	namein, err := textinput.New(
		textinput.Label("NAME ", cell.Bold(), cell.FgColor(cell.ColorNumber(33))),
		textinput.TextColor(cell.ColorWhite),
		textinput.MaxWidthCells(35),
		textinput.ExclusiveKeyboardOnFocus(),
	)

	cont, err := container.New(
		t,
		container.ID("configs_display"),
		container.Border(linestyle.Light),
		container.BorderTitle("CONFIGS (ESC to go back)"),
		container.KeyFocusNext(keyboard.KeyTab),
		container.KeyFocusGroupsNext(keyboard.KeyArrowDown, 1),
		container.KeyFocusGroupsPrevious(keyboard.KeyArrowUp, 1),
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.BorderTitle("Log"),
						container.PlaceWidget(errlog),
					),
					container.Right(
						container.SplitVertical(
							container.Left(
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
										),
									),
								),
							),
							container.Right(),
							container.SplitPercent(80),
						),
					),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Configurated"),
				container.SplitVertical(
					container.Left(),
					container.Right(),
				),
			),
			container.SplitPercent(18),
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
			dbms = dbmsin.ReadAndClear()
			if utility.Fstr(dbms) == "" || utility.Fstr(dbms) != "MYSQL" {
				inerr = "Unknown DBMS: " + dbms
				insuc = ""
				dsn = dsnin.ReadAndClear()
				name = namein.ReadAndClear()
				break
			}

			dsn = dsnin.ReadAndClear()
			if string(dsn) == "" {
				inerr = "Blank DSN is invalid."
				insuc = ""
				name = namein.ReadAndClear()
				break
			}

			name = namein.ReadAndClear()
			if name == "" {
				name = "null"
			}

			var inst types.Instance
			inst.DBMS = utility.Dbmsstr(dbms)
			inst.DSN = []byte(dsn)
			inst.Dbname = name

			if !db.Ping(inst) {
				inerr = "Invalid DSN. Connection closed."
				insuc = ""
				break
			} else {
				insuc = "Connection established!"
				inerr = ""
			}

			instances = append(instances, inst)
			instances = io.SyncConfig(instances)

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
*/
func DisplayDbDashboard(t *tcell.Terminal, cpool []*sql.DB) {
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
	bfptext, _ := text.New()

	go dynDbDashboard(ctx, infotext, bfptext, Interval, cpool)

	cont, err := container.New(
		t,
		container.ID("db_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("INNODB DASHBOARD (? for help)"),
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
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
			container.Right(),
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

func DisplayMemory(t *tcell.Terminal, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("memory_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("MEMORY (? for help)"),
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

func DisplayErrorLog(t *tcell.Terminal, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("err_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("ERROR LOG (? for help)"),
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

func DisplayLocks(t *tcell.Terminal, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("lock_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("LOCKS (? for help)"),
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
