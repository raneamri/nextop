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
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/queries"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

func DisplayMemory() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	go connectionSanitiser(ctx, cancel)

	dballoc_txt, _ := text.New()
	dballoc_txt.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	usralloc_txt, _ := text.New()
	usralloc_txt.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	dballoc_lc, _ := linechart.New(
		linechart.YAxisAdaptive(),
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorOlive)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorOlive)),
	)

	hardwalloc_txt, _ := text.New()
	hardwalloc_txt.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynMemoryDashboard(ctx,
		dballoc_txt,
		usralloc_txt,
		dballoc_lc,
		hardwalloc_txt)

	cont, err := container.New(
		t,
		container.ID("memory_dashboard"),
		container.Border(linestyle.Light),
		container.BorderTitle("MEMORY (? for help, +/- to change refresh rate)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Db Memory Allocation"),
						container.PlaceWidget(dballoc_txt),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Users Memory Allocation"),
						container.PlaceWidget(usralloc_txt),
					),
					container.SplitPercent(65),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle(fmt.Sprint("Total Allocated Memory")),
						container.PlaceWidget(dballoc_lc),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Disk / RAM"),
						container.PlaceWidget(hardwalloc_txt),
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
		case '?':
			State = types.MENU
			cancel()
		case '+':
			Interval += 100 * time.Millisecond
			cancel()
		case '-':
			if Interval > 100*time.Millisecond {
				Interval -= 100 * time.Millisecond
			}
			cancel()
		case keyboard.KeyTab:
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
	dballoc_txt *text.Text,
	usralloc_txt *text.Text,
	linechart *linechart.LineChart,
	hardwalloc_txt *text.Text) {

	var (
		mallocChannel      chan types.Query = make(chan types.Query)
		globalallocChannel chan types.Query = make(chan types.Query)
		ramndiskChannel    chan types.Query = make(chan types.Query)
		specallocChannel   chan types.Query = make(chan types.Query)
	)

	go fetchMemoryAlloc(ctx,
		mallocChannel,
		globalallocChannel,
		specallocChannel,
		ramndiskChannel)

	go displayMemoryAlloc(ctx,
		mallocChannel,
		globalallocChannel,
		ramndiskChannel,
		specallocChannel,
		dballoc_txt,
		usralloc_txt,
		hardwalloc_txt,
		linechart)

	<-ctx.Done()
}

func fetchMemoryAlloc(ctx context.Context,
	mallocChannel chan<- types.Query,
	globalallocChannel chan<- types.Query,
	specallocChannel chan<- types.Query,
	ramndiskChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string

		malloc_query      types.Query
		globalalloc_query types.Query
		ramndisk_query    types.Query
		specalloc_query   types.Query
	)

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]

			malloc_query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["malloc"]())
			globalalloc_query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["globalalloc"]())
			ramndisk_query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["ramndisk"]())
			specalloc_query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["specalloc"]())

			mallocChannel <- malloc_query
			globalallocChannel <- globalalloc_query
			ramndiskChannel <- ramndisk_query
			specallocChannel <- specalloc_query

		case <-ctx.Done():
			return
		}
	}

}

func displayMemoryAlloc(ctx context.Context,
	mallocChannel <-chan types.Query,
	globalallocChannel <-chan types.Query,
	ramndiskChannel <-chan types.Query,
	specallocChannel <-chan types.Query,
	widget_db *text.Text,
	widget_usr *text.Text,
	widget_hardw *text.Text,
	lc *linechart.LineChart) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query types.Query

		malloc_text_buffer      [][]string = make([][]string, 0)
		globalalloc_text_buffer [][]string = make([][]string, 0)
		specalloc_text_buffer   [][]string = make([][]string, 0)
		ramndisk_text_buffer    [][]string = make([][]string, 0)

		lc_globalalloc []float64

		format      string = "\n  %-30v %17v %15v"
		format_pair string = "\n  %-30v %17v"
	)

	for {
		select {
		case query = <-mallocChannel:
			for _, row := range query.RawData {
				malloc_text_buffer = append(malloc_text_buffer, row)
			}

		case query = <-globalallocChannel:
			for _, row := range query.RawData {
				globalalloc_text_buffer = append(globalalloc_text_buffer, row)
			}

		case query = <-specallocChannel:
			for _, row := range query.RawData {
				specalloc_text_buffer = append(specalloc_text_buffer, row)
			}

		case query = <-ramndiskChannel:
			for _, row := range query.RawData {
				ramndisk_text_buffer = append(ramndisk_text_buffer, row)
			}

		case <-ticker.C:
			widget_db.Reset()
			widget_usr.Reset()
			widget_hardw.Reset()

			for _, row := range malloc_text_buffer {
				for i := range row {
					if i < len(row)-2 {
						widget_usr.Write(utility.TrimNSprintf(format, row[i], row[i+1], row[i+2]))
					}
				}
			}

			for _, row := range specalloc_text_buffer {
				for i := range row {
					if i < len(row)-1 {
						widget_db.Write(utility.TrimNSprintf(format_pair, row[i], row[i+1]))
					}
				}
			}

			for _, row := range ramndisk_text_buffer {
				for i := range row {
					if i < len(row)-2 {
						widget_hardw.Write(utility.TrimNSprintf(format, row[i], row[i+1], row[i+2]))
					}
				}
			}

			malloc_text_buffer = make([][]string, 0)
			specalloc_text_buffer = make([][]string, 0)
			ramndisk_text_buffer = make([][]string, 0)

			parts := strings.SplitN(globalalloc_text_buffer[0][0], " ", 2)
			aps, _ := strconv.ParseFloat(parts[0], 64)
			lc_globalalloc = append(lc_globalalloc, aps)

			if err := lc.Series("Errors", lc_globalalloc,
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
