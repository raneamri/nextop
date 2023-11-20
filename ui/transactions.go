package ui

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
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

func DisplayTransactions() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		pause atomic.Value
	)
	pause.Store(false)

	txns_text, _ := text.New()
	txns_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	info_text, _ := text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynTransactions(ctx,
		txns_text,
		info_text,
		&pause)

	cont, err := container.New(
		t,
		container.ID("transactions"),
		container.Border(linestyle.Light),
		container.BorderTitle("TRANSACTIONS (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Transactions"),
				container.PlaceWidget(txns_text),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle(""),
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
		case '?':
			State = types.MENU
			cancel()
		case '=':
			if pause.Load().(bool) {
				pause.Store(false)
			} else {
				pause.Store(true)
			}
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

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}

func dynTransactions(ctx context.Context,
	txns_text *text.Text,
	info_text *text.Text,
	pause *atomic.Value) {

	var (
		txnsChannel    chan types.Query = make(chan types.Query)
		metricsChannel chan types.Query = make(chan types.Query)
	)

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

	go fetchTransactions(ctx, txnsChannel, pause)
	go writeTransactions(ctx, txns_text, txnsChannel)

	<-ctx.Done()
}

func fetchTransactions(ctx context.Context,
	txnsChannel chan<- types.Query,
	pause *atomic.Value) {

	var (
		ticker *time.Ticker = time.NewTicker(1 * time.Nanosecond)
		istIte bool         = false

		lookup map[string]func() string

		query types.Query

		order []int = make([]int, 10)
	)

	switch Instances[ActiveConns[0]].DBMS {
	case types.MYSQL:
		order = []int{0, 1, 2, 3, 4}

	case types.POSTGRES:
		order = nil

	default:
		order = nil
	}

	for {
		select {
		case <-ticker.C:
			if !istIte {
				istIte = true
				ticker = time.NewTicker(Interval)
				defer ticker.Stop()
			}

			if pause.Load().(bool) {
				continue
			}
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["transactions"]())

			utility.ShuffleQuery(query, order)

			txnsChannel <- query

		case <-ctx.Done():
			return
		}
	}

}

func writeTransactions(ctx context.Context,
	widget *text.Text,
	txnsChannel <-chan types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		header []interface{} = []interface{}{"Thd", "User", "Cmd", "Duration", "Stmt"}
		format string        = "%-9v %-25v %-15v %-15v %-64v\n"

		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case query = <-txnsChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()

			widget.Write(fmt.Sprintf(format, header...), text.WriteCellOpts(cell.Bold()))

			colorflipper = 1

			for _, row := range text_buffer {
				if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				}
				colorflipper *= -1

				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...), color)
			}

			text_buffer = make([][]string, 0)

		case <-ctx.Done():
			return
		}
	}
}
