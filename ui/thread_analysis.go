package ui

import (
	"context"
	"fmt"
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
	"github.com/mum4k/termdash/widgets/text"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

var CurrentThread string
var CurrentQuery string
var CurrentConn string

func DisplayThreadAnalysis() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	go connectionSanitiser(ctx, cancel)

	var (
		lookup  map[string]func() string
		message []string
		query   types.Query
		abort   bool = false

		header []interface{} = []interface{}{"ID", "QType", "Table", "Parts", "Type", "Key", "Key Len", "Ref", "Rows", "Filtered", "Extra"}
		format string        = "%-10v %-10v %-10v %-10v %-10v %-10v %-10v %-10v %-10v %-10v %-10v \n"

		order []int = []int{0, 1, 2, 3, 4, 6, 7, 8, 9, 10, 11}

		color        text.WriteOption
		colorflipper int
	)

	widget, _ := text.New()

	_, err = strconv.Atoi(CurrentThread)
	if err != nil {
		message = append(message, "Non-numeric Thread ID\nESC to return to previous view")
		abort = true
	} else {
		lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
		query, _ = queries.Query(Instances[ActiveConns[0]].Driver, fmt.Sprintf(lookup["thread_analysis"](), CurrentQuery))
		utility.ShuffleQuery(query, order)
	}

	CurrentQuery = ""

	widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

	if abort {
		widget.Write(message[0], text.WriteCellOpts(cell.FgColor(cell.ColorYellow)))
	} else {
		for _, row := range query.RawData {
			if colorflipper < 0 {
				color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
			} else {
				color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
			}
			colorflipper *= -1

			widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...), color)
		}
	}

	cont, err := container.New(
		t,
		container.ID("thread_analysis"),
		container.Border(linestyle.Light),
		container.BorderTitle("THREAD ANALYSIS (ESC to return)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.PlaceWidget(widget),
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
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyTab:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}
