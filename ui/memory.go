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
	"github.com/raneamri/nextop/db"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:

*/

/*
Format:

	widget-1 (top-left): general alloc
	widget-2 (bottom-left): user alloc
	widget-3 (top-right): alloc/time
	widget-4 (bottom-right): disk&ram/alloc
*/
func DisplayMemory() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	dballoc_labels, _ := text.New()
	dballoc_txt, _ := text.New()
	dballoc_txt.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	usralloc_labels, _ := text.New()
	usralloc_txt, _ := text.New()
	usralloc_txt.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	dballoc_lc, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	hardwalloc_labels, _ := text.New()
	hardwalloc_txt, _ := text.New()
	hardwalloc_txt.Write("\n\nLoading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		Slice to hold allocated memory over time
	*/
	var alt []float64

	go dynMemoryDashboard(ctx,
		dballoc_labels, dballoc_txt,
		usralloc_labels, usralloc_txt,
		dballoc_lc,
		hardwalloc_labels, hardwalloc_txt,
		alt, Interval)

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
								container.PlaceWidget(dballoc_labels),
							),
							container.Right(
								container.PlaceWidget(dballoc_txt),
							),
							container.SplitPercent(60),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Users Memory Allocation"),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(usralloc_labels),
							),
							container.Right(
								container.PlaceWidget(usralloc_txt),
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
						container.BorderTitle("Disk / RAM Allocation"),
						container.SplitVertical(
							container.Left(
								container.PlaceWidget(hardwalloc_labels),
							),
							container.Right(
								container.PlaceWidget(hardwalloc_txt),
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
		elapsed := time.Since(LastInputTime)
		ratelim, _ := strconv.Atoi(io.FetchSetting("rate-limiter"))
		if elapsed < time.Duration(ratelim)*time.Millisecond {
			return
		}
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
	dballoc_labels *text.Text, dballoc_txt *text.Text,
	usralloc_labels *text.Text, usralloc_txt *text.Text,
	dballoc_lc *linechart.LineChart,
	hardwalloc_labels *text.Text, hardwalloc_txt *text.Text,
	alt []float64,
	delay time.Duration) {

	var (
		dballocChannel    chan [2]string = make(chan [2]string)
		usrallocChannel   chan [2]string = make(chan [2]string)
		lcChannel         chan []float64 = make(chan []float64)
		hardwallocChannel chan [2]string = make(chan [2]string)
	)

	go fetchMemoryDbAlloc(ctx, dballocChannel, lcChannel, delay)
	go writeMemoryDbAlloc(ctx, dballoc_labels, dballoc_txt, dballoc_lc, dballocChannel, lcChannel)

	go fetchMemoryUserAlloc(ctx, usrallocChannel, delay)
	go writeMemoryUserAlloc(ctx, usralloc_labels, usralloc_txt, usrallocChannel)

	go fetchMemoryHardwAlloc(ctx, hardwallocChannel, delay)
	go writeMemoryHardwAlloc(ctx, hardwalloc_labels, hardwalloc_txt, hardwallocChannel)

	<-ctx.Done()
}

/*
widgets-1 & 3
*/
func fetchMemoryDbAlloc(ctx context.Context, dballocChannel chan<- [2]string, lcChannel chan<- []float64, delay time.Duration) {
	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup       map[string]func() string = make(map[string]func() string)
		global_alloc [][]string               = make([][]string, 0)
		parts        []string                 = make([]string, 0)
		aps          float64

		/*
			Channel message variable
		*/
		messages    [2]string
		lc_messages []float64 = make([]float64, 0)
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[CurrConn].DBMS]
			global_alloc = db.GetLongQuery(Instances[CurrConn].Driver, lookup["global_alloc"]())
			alloc_by_area := db.GetLongQuery(Instances[CurrConn].Driver, lookup["spec_alloc"]())

			messages[0] += "\n\n   Total allocated\n\n"
			messages[1] += "\n\n" + global_alloc[0][0] + "\n\n"
			for _, chunk := range alloc_by_area {
				messages[0] += "   " + strings.TrimLeft(chunk[0], " ") + "\n"
				messages[1] += strings.TrimLeft(chunk[1], " ") + "\n"
			}

			parts = strings.SplitN(global_alloc[0][0], " ", 2)
			aps, _ = strconv.ParseFloat(parts[0], 64)
			lc_messages = append(lc_messages, aps)

			dballocChannel <- messages
			lcChannel <- lc_messages

			messages = [2]string{""}
		case <-ctx.Done():
			return
		}
	}
}

func writeMemoryDbAlloc(ctx context.Context,
	dballoc_labels *text.Text, dballoc_txt *text.Text,
	dballoc_lc *linechart.LineChart,
	dballocChannel <-chan [2]string, lcChannel <-chan []float64) {

	var (
		/*
			Display variables
		*/
		messages    [2]string
		lc_messages []float64 = make([]float64, 0)
	)

	for {
		select {
		/*
			No need to check both channels since we know
			lcChannel updates right before dballocChannel
		*/
		case messages = <-dballocChannel:
			dballoc_labels.Reset()
			dballoc_txt.Reset()

			dballoc_labels.Write(messages[0], text.WriteCellOpts(cell.Bold()))
			dballoc_txt.Write(messages[1])

			lc_messages = <-lcChannel
			if err := dballoc_lc.Series("first", lc_messages,
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

/*
widget-2
*/
func fetchMemoryUserAlloc(ctx context.Context, usrallocChannel chan<- [2]string, delay time.Duration) {
	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup    map[string]func() string = make(map[string]func() string)
		usr_alloc [][]string               = make([][]string, 0)

		/*
			Channel message variable
		*/
		messages [2]string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[CurrConn].DBMS]
			usr_alloc = db.GetLongQuery(Instances[CurrConn].Driver, lookup["user_alloc"]())

			messages[0] += "\n\n   User\n\n"
			messages[1] += fmt.Sprintf("\n\n%-11v %-9v\n", "Current", "(Max)")
			for _, chunk := range usr_alloc {
				messages[0] += "   " + chunk[0] + "\n"
				messages[1] += fmt.Sprintf("\n%-11v %-9v", strings.TrimLeft(chunk[1], " "), "("+strings.ReplaceAll(chunk[2], " ", "")+")")
			}

			usrallocChannel <- messages
			messages = [2]string{""}
		case <-ctx.Done():
			return
		}
	}
}

func writeMemoryUserAlloc(ctx context.Context, usralloc_labels *text.Text, usralloc_txt *text.Text, usrallocChannel <-chan [2]string) {
	var (
		/*
			Display variables
		*/
		messages [2]string
	)

	for {
		select {
		case messages = <-usrallocChannel:
			usralloc_labels.Reset()
			usralloc_txt.Reset()

			usralloc_labels.Write(messages[0], text.WriteCellOpts(cell.Bold()))
			usralloc_txt.Write(messages[1])

		case <-ctx.Done():
			return
		}
	}
}

/*
widget-4
*/
func fetchMemoryHardwAlloc(ctx context.Context, hardwallocChannel chan<- [2]string, delay time.Duration) {
	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup         map[string]func() string = make(map[string]func() string)
		ramndisk_alloc [][]string               = make([][]string, 0)

		/*
			Channel message variable
		*/
		messages [2]string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[CurrConn].DBMS]
			ramndisk_alloc = db.GetLongQuery(Instances[CurrConn].Driver, lookup["ramdisk_alloc"]())

			messages[0] += "\n\n\n\n                   Disk\n                    RAM"
			messages[1] += fmt.Sprintf("\n\n%-11v %-9v\n", "Current", "(Max)")
			for _, chunk := range ramndisk_alloc {
				messages[1] += fmt.Sprintf("\n%-11v %-9v", strings.TrimLeft(chunk[1], " "), "("+strings.ReplaceAll(chunk[2], " ", "")+")")
			}

			hardwallocChannel <- messages
			messages = [2]string{""}

		case <-ctx.Done():
			return
		}
	}
}

func writeMemoryHardwAlloc(ctx context.Context, hardwalloc_labels *text.Text, hardwalloc_txt *text.Text, hardwallocChannel <-chan [2]string) {
	var (
		/*
			Display variables
		*/
		messages [2]string
	)

	for {
		select {
		case messages = <-hardwallocChannel:
			hardwalloc_labels.Reset()
			hardwalloc_txt.Reset()

			hardwalloc_labels.Write(messages[0], text.WriteCellOpts(cell.Bold()))
			hardwalloc_txt.Write(messages[1])

		case <-ctx.Done():
			return
		}
	}
}
