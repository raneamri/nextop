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

	txns_text, _ := text.New()

	go dynTransactions(ctx,
		txns_text,
		Interval)

	cont, err := container.New(
		t,
		container.ID("transactions"),
		container.Border(linestyle.Light),
		container.BorderTitle("TRANSACTIONS (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.PlaceWidget(txns_text),
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
			CurrRotateLeft()
		case keyboard.KeyArrowRight:
			CurrRotateRight()
		case '?':
			State = types.MENU
			cancel()
		case keyboard.KeyCtrlD:
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

func dynTransactions(ctx context.Context,
	txns_text *text.Text,
	delay time.Duration) {

	var (
		txnsChannel chan [][]string = make(chan [][]string)
	)

	go fetchTransactions(ctx, txnsChannel, delay)
	go writeTransactions(ctx, txns_text, txnsChannel)

	<-ctx.Done()
}

func fetchTransactions(ctx context.Context,
	txnsChannel chan<- [][]string,
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
			lookup = GlobalQueryMap[Instances[CurrConn].DBMS]
			messages = queries.GetLongQuery(Instances[CurrConn].Driver, lookup["transactions"]())

			txnsChannel <- messages

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
		header  string
		message [][]string = make([][]string, 0)
	)

	for {
		select {
		case message = <-txnsChannel:
			txns_text.Reset()

			header = fmt.Sprintf("%-5v %-15v\n", "H1", "H2")

			txns_text.Write(header)

			for _, line := range message {
				for _, item := range line {
					txns_text.Write(item)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
