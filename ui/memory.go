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

func dynMemoryDashboard(ctx context.Context,
	dballoc1_txt *text.Text,
	dballoc2_txt *text.Text,
	usralloc1_txt *text.Text,
	usralloc2_txt *text.Text,
	dballoc_lc *linechart.LineChart,
	hardwalloc1_txt *text.Text,
	hardwalloc2_txt *text.Text,
	alt []float64,
	delay time.Duration) {

	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			lookup := GlobalQueryMap[Instances[CurrConn].DBMS]
			total_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["global_alloc"]())
			alloc_by_area := db.GetLongQuery(Instances[CurrConn].Driver, lookup["spec_alloc"]())
			usr_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["user_alloc"]())
			ramndisk_alloc := db.GetLongQuery(Instances[CurrConn].Driver, lookup["ramdisk_alloc"]())

			dballoc1_txt.Reset()
			dballoc2_txt.Reset()
			dballoc1_txt.Write("\n\n   Total allocated\n\n", text.WriteCellOpts(cell.Bold()))
			dballoc2_txt.Write("\n\n"+total_alloc[0][0]+"\n\n", text.WriteCellOpts(cell.Bold()))
			for i, chunk := range alloc_by_area {
				dballoc1_txt.Write("   "+strings.TrimLeft(chunk[0], " ")+"\n", text.WriteCellOpts(cell.Bold()))
				if i%2 == 0 {
					dballoc2_txt.Write(strings.TrimLeft(chunk[1], " ")+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					dballoc2_txt.Write(strings.TrimLeft(chunk[1], " ")+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			usralloc1_txt.Reset()
			usralloc2_txt.Reset()
			usralloc1_txt.Write("\n\n   User\n\n", text.WriteCellOpts(cell.Bold()))
			usralloc2_txt.Write("\n\nCurrent  (Max)\n\n", text.WriteCellOpts(cell.Bold()))
			for j, chunk := range usr_alloc {
				usralloc1_txt.Write("   "+chunk[0]+"\n", text.WriteCellOpts(cell.Bold()))
				chunk[2] = strings.ReplaceAll(chunk[2], " ", "")
				if j%2 == 0 {
					usralloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					usralloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			hardwalloc1_txt.Reset()
			hardwalloc2_txt.Reset()
			hardwalloc1_txt.Write("\n\n\n\n                   Disk\n                    RAM", text.WriteCellOpts(cell.Bold()))
			hardwalloc2_txt.Write("\n\nCurrent  (Max)\n\n", text.WriteCellOpts(cell.Bold()))
			for k, chunk := range ramndisk_alloc {
				chunk[2] = strings.ReplaceAll(chunk[2], " ", "")
				if k%2 == 0 {
					hardwalloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorGray)))
				} else {
					hardwalloc2_txt.Write(strings.TrimLeft(chunk[1], " ")+" ("+chunk[2]+")"+"\n", text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
				}
			}

			parts := strings.SplitN(total_alloc[0][0], " ", 2)
			aps, _ := strconv.ParseFloat(parts[0], 64)
			if aps > 0 {
				alt = append(alt, aps)
			}
			if err := dballoc_lc.Series("first", alt,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(33))),
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
