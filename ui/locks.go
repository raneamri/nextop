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
	"github.com/raneamri/nextop/utility"

	_ "github.com/go-sql-driver/mysql"
)

func DisplayLocks() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	active_txt, _ := text.New()
	log, _ := text.New()

	go dynLockLog(ctx, log)

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
	log *text.Text) {

	var (
		lockChannel chan types.Query = make(chan types.Query)
	)

	go fetchLocks(ctx, lockChannel)
	go writeLocks(ctx, log, lockChannel)
}

func fetchLocks(ctx context.Context,
	lockChannel chan<- types.Query) {

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
			query, _ = queries.Query(Instances[ActiveConns[0]].Driver, lookup["locks"]())

			lockChannel <- query

		case <-ctx.Done():
			return
		}
	}
}

func writeLocks(ctx context.Context,
	widget *text.Text,
	lockChannel <-chan types.Query) {

	var (
		ticker *time.Ticker = time.NewTicker(Interval)

		query       types.Query
		text_buffer [][]string = make([][]string, 0)

		//header []interface{} = []interface{}{"Timestamp", "Thd", "Type", "Message"}
		format string = "%-200v\n"

		color        text.WriteOption
		colorflipper int = -1
	)

	for {
		select {
		case query = <-lockChannel:
			for _, row := range query.RawData {
				text_buffer = append(text_buffer, row)
			}

		case <-ticker.C:
			widget.Reset()

			for _, row := range text_buffer {
				if colorflipper > 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				}
				colorflipper *= -1

				widget.Write(utility.TrimNSprintf(format, utility.SliceToInterface(row)...), color)
			}

		case <-ctx.Done():
			return
		}
	}
}
