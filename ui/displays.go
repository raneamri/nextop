package ui

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/raneamri/gotop/db"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
Contains all displays that don't involve i/o (excluding keyboard reads)
*/

/*
Draws the main menu and reads keyboard input
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
		case '?':
			State = types.HELP
			cancel()
		case 'b', 'B':
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
*/
func DrawHelp(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("help_screen"),
		container.Border(linestyle.Light),
		container.BorderTitle("GOTOP/HELP (b to go back)"),
	)
	if err != nil {
		panic(err)
	}

	/*
		Note: add help screen
	*/

	quitter := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case 'p', 'P':
			State = types.PROCESSLIST
			cancel()
		case 'b', 'B':
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}

}

/*
Format of this display is:

	container-1 (bottom): holds parsed processlist data
	container-2 (top-left): shows uptime & qps
	container-3 (top-right): shows lifeline as graph
*/
func DisplayProcesslist(t *tcell.Terminal, cpool []*sql.DB) {
	ctx, cancel := context.WithCancel(context.Background())

	pl_header := fmt.Sprintf("%-7v %-5v %-5v %-7v %-25v %-20v %-12v %10v %10v %-65v\n",
		"Cmd", "Thd", "Conn", "PID", "State", "User", "Db", "Time", "Lock Time", "Query")
	tl_header := fmt.Sprintf("%-15v %-20v %-5v\n",
		"Uptime", "QPS", "Threads")

	_, data, _ := db.GetProcesslist(cpool[0])
	uptime := db.GetUptime(cpool[0])
	qps := db.GetUptime(cpool[0])

	tl_table, _ := text.New()
	if err := tl_table.Write(tl_header, text.WriteCellOpts(cell.Bold())); err != nil {
		panic(err)
	}
	frow := fmt.Sprintf("%-15v %-20v %-5v", utility.Ftime(uptime), fmt.Sprint(qps), " ")
	tl_table.Write(frow, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))

	pl_table, _ := text.New()
	if err := pl_table.Write(pl_header, text.WriteCellOpts(cell.Bold()), text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
		panic(err)
	}

	flip := 1
	for _, row := range data {
		ftime, _ := strconv.ParseInt(row[8], 10, 64)
		row[8] = utility.FpicoToMs(ftime)
		flocktime, _ := strconv.ParseInt(row[9], 10, 64)
		row[9] = utility.FpicoToUs(flocktime)
		frow := fmt.Sprintf("%-7v %-5v %-5v %-7v %-25v %-20v %-12v %10v %10v %-65v\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[8], row[9], row[7])
		if flip > 0 {
			pl_table.Write(frow, text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
		} else if flip < 0 {
			pl_table.Write(frow, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
		}
		flip *= -1
	}

	cont, err := container.New(
		t,
		container.ID("processlist"),
		container.Border(linestyle.Light),
		container.BorderTitle("GOTOP/PROCESSLIST (? for help, b to go back, ,<- -> to navigate)"),
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
							),
							container.Right(
								container.Border(linestyle.Light),
								container.BorderTitle("Health"),
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
				container.PlaceWidget(pl_table),
			),
			container.SplitPercent(40),
		),
	)

	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case 'b', 'B':
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}
