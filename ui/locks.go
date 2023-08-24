package ui

import (
	"context"
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

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:

	1 query
	across 3 goroutines
*/

/*
Format:

	widget-1 (top): blank
	widget-2 (bottom): locks
*/
func DisplayLocks() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	active_txt, _ := text.New()
	log, _ := text.New()

	go dynLockLog(ctx, log, Interval)

	cont, err := container.New(
		t,
		container.ID("lock_log"),
		container.Border(linestyle.Light),
		container.BorderTitle("LOCKS LOG (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.PlaceWidget(active_txt),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("Locks"),
				container.PlaceWidget(log),
			),
			container.SplitPercent(20),
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
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case keyboard.KeyTab:
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

func dynLockLog(ctx context.Context,
	log *text.Text,
	delay time.Duration) {

	var (
		lockChannel chan string = make(chan string)
	)

	go fetchLocks(ctx, lockChannel, delay)
	go writeLocks(ctx, log, lockChannel)
}

func fetchLocks(ctx context.Context, lockChannel chan<- string, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup map[string]func() string

		messages [][]string
		message  string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			messages = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["locks"]())

			if len(messages) == 0 {
				message = "No active locks"
			} else {
				message = messages[0][0]
			}

			lockChannel <- message

		case <-ctx.Done():
			return
		}
	}
}

func writeLocks(ctx context.Context, log *text.Text, lockChannel <-chan string) {

	var (
		message string
	)

	for {
		select {
		case message = <-lockChannel:

			log.Reset()
			log.Write(message)

		case <-ctx.Done():
			return
		}
	}
}
