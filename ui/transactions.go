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
)

/*
Workload:
*/

/*
Formats:
*/

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

	go dynTransactions(ctx,
		txns_text,
		&pause,
		Interval)

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
			),
			container.SplitPercent(90),
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
	pause *atomic.Value,
	delay time.Duration) {

	var (
		txnsChannel chan [][]string = make(chan [][]string)
	)

	go fetchTransactions(ctx, txnsChannel, pause, delay)
	go writeTransactions(ctx, txns_text, txnsChannel)

	<-ctx.Done()
}

func fetchTransactions(ctx context.Context,
	txnsChannel chan<- [][]string,
	pause *atomic.Value,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		/*
			Fetch variables
		*/
		lookup map[string]func() string
		/*
			Formatting variables
		*/

		/*
			Channel message variable
		*/
		messages [][]string = make([][]string, 0)
	)

	for {
		select {
		case <-ticker.C:
			if pause.Load().(bool) {
				continue
			}
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			messages = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["transactions"]())

			if len(messages) < 1 {
				messages[0][0] = "No active transactions"
			}

			txnsChannel <- messages
			messages = [][]string{}

		case <-ctx.Done():
			return
		}
	}

}

func writeTransactions(ctx context.Context,
	txns_text *text.Text,
	txnsChannel <-chan [][]string) {

	var (
		/*
			Parse variables
		*/

		/*
			Display variables
		*/
		color        text.WriteOption
		colorflipper int = 1
		header       string
		message      [][]string = make([][]string, 0)
		out          string
	)

	for {
		select {
		case message = <-txnsChannel:
			txns_text.Reset()

			header = fmt.Sprintf("%-9v %-25v %-15v %-15v %-64v\n", "Thd", "User", "Cmd", "Duration", "Stmt")

			txns_text.Write(header, text.WriteCellOpts(cell.Bold()))

			for _, line := range message {
				if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				}
				colorflipper *= -1
				out = fmt.Sprintf("%-9v %-25v %-15v %-15v %-64v\n", line[0], line[1], line[2], line[3], line[4])
				txns_text.Write(out, color)
			}

		case <-ctx.Done():
			return
		}
	}
}
