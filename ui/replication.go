package ui

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
)

func DisplayReplication() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	go connectionSanitiser(ctx, cancel)

	var (
		repl_text *text.Text
		info_text *text.Text
	)

	repl_text, _ = text.New()
	repl_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	info_text, _ = text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynReplication(ctx,
		repl_text,
		info_text)

	cont, err := container.New(
		t,
		container.ID("replication"),
		container.Border(linestyle.Light),
		container.BorderTitle("REPLICATION (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Replication Status"),
				container.BorderColor(cell.ColorGray),
				container.FocusedColor(cell.ColorWhite),
				container.PlaceWidget(repl_text),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Connections (<-/-> to swap)"),
				container.BorderColor(cell.ColorGray),
				container.FocusedColor(cell.ColorWhite),
				container.PlaceWidget(info_text),
			),
			container.SplitPercent(70),
		),
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
		case keyboard.KeyArrowLeft:
			RotateConnsLeft()
		case keyboard.KeyArrowRight:
			RotateConnsRight()
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyTab:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case '+':
			Interval += 100 * time.Millisecond
		case '-':
			if Interval > 100*time.Millisecond {
				Interval -= 100 * time.Millisecond
			}
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

func dynReplication(ctx context.Context,
	repl_text *text.Text,
	info_text *text.Text) {

	var (
		replicationChannel chan types.Query = make(chan types.Query)
		metricsChannel     chan types.Query = make(chan types.Query)
	)

	defer close(replicationChannel)
	defer close(metricsChannel)

	for _, conn := range ActiveConns {
		go fetchMetrics(ctx,
			conn,
			metricsChannel)
	}

	go displayMetrics(ctx,
		metricsChannel,
		info_text,
		nil,
		nil)

	go fetchReplication(ctx, replicationChannel)
	go writeReplication(ctx, repl_text, replicationChannel)

	<-ctx.Done()
}

func fetchReplication(ctx context.Context,
	replicationChannel chan<- types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string

		query types.Query
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
			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["replication"]())

			if len(query.RawData) == 0 {
				query.RawData = [][]string{{"NO CURR. REPL", "", "", "", "", "", "", ""}}
			}

			replicationChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func writeReplication(ctx context.Context,
	widget *text.Text,
	replicationChannel <-chan types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		header []interface{} = []interface{}{"Master Host", "Master Port", "Master Log File",
			"Read Master Log Pos", "Slave IO Running", "Slave SQL Running", "Secs Behind Master",
			"Last Error"}
		format string = "%-15v %-16v %-18v %-22v %-24v %-20v %-20v %-20v\n"
	)

	for {
		select {
		case query = <-replicationChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()
			widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

			for _, row := range text_buffer {
				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...))
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}
