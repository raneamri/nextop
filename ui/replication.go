package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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

/*
Workload:
*/

/*
Formats:
*/

func DisplayReplication() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		repl_text *text.Text
		info_text *text.Text
	)

	/*
		widget-1 (top)
	*/
	repl_text, _ = text.New()
	repl_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	/*
		widget-2 (bottom)
	*/
	info_text, _ = text.New()
	info_text.Write("Loading...", text.WriteCellOpts(cell.FgColor(cell.ColorNavy)))

	go dynReplication(ctx,
		repl_text,
		info_text,
		Interval)

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
			container.SplitPercent(35),
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
	info_text *text.Text,
	delay time.Duration) {

	var (
		replicationChannel chan []string = make(chan []string)
		infoChannel        chan []string = make(chan []string)
	)

	go fetchReplication(ctx, replicationChannel, delay)
	go writeReplication(ctx, repl_text, replicationChannel)

	go fetchInfo(ctx, infoChannel, delay)
	go writeInfo(ctx, info_text, infoChannel)

	<-ctx.Done()
}

func fetchReplication(ctx context.Context, replicationChannel chan<- []string, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup map[string]func() string

		messages [][]string
		message  []string
	)

	for {
		select {
		case <-ticker.C:
			lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
			messages = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["replication"]())

			if len(messages) == 0 {
				message = []string{"No ongoing replication. Swap connections in the view below using <-/->."}
			} else {
				message = messages[0]
			}

			replicationChannel <- message

		case <-ctx.Done():
			return
		}
	}
}

func writeReplication(ctx context.Context, log *text.Text, replicationChannel <-chan []string) {

	var (
		message []string
		headers string
	)

	for {
		select {
		case message = <-replicationChannel:

			headers = fmt.Sprintf("%-15v %-16v %-18v %-22v %-24v %-20v %-20v %-20v\n", "Master Host", "Master Port", "Master Log File",
				"Read Master Log Pos", "Slave IO Running", "Slave SQL Running", "Secs Behind Master",
				"Last Error")

			log.Reset()
			log.Write(headers, text.WriteCellOpts(cell.Bold()))
			for _, msg := range message {
				log.Write(msg)
			}

		case <-ctx.Done():
			return
		}
	}
}

/*
container-2
*/
func fetchInfo(ctx context.Context,
	infoChannel chan<- []string,
	delay time.Duration) {

	var ticker *time.Ticker = time.NewTicker(delay)
	defer ticker.Stop()

	var (
		lookup map[string]func() string

		statuses [][]string = make([][]string, 0)
		qps_int  int
		uptime   int

		messages []string = make([]string, 0)
	)

	for {
		select {
		case <-ticker.C:
			for _, key := range ActiveConns {
				lookup = GlobalQueryMap[Instances[key].DBMS]
				statuses = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, lookup["uptime"]())

				uptime, _ = strconv.Atoi(statuses[1][1])
				qps_int, _ = strconv.Atoi(queries.GetLongQuery(Instances[key].Driver, lookup["queries"]())[0][0])

				messages = append(messages, fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
					key, utility.Ftime(uptime), utility.Fnum(qps_int), statuses[0][1]))
			}

			infoChannel <- messages
			messages = []string{}

		case <-ctx.Done():
			return
		}
	}
}

func writeInfo(ctx context.Context,
	info_text *text.Text,
	infoChannel <-chan []string) {

	var (
		message      []string = make([]string, 0)
		color        text.WriteOption
		colorflipper int
	)

	for {
		select {
		case message = <-infoChannel:
			headers := fmt.Sprintf("%-13v %-22v %-10v %-5v\n",
				"Connection", "Uptime", "Queries", "Threads")

			colorflipper = -1

			info_text.Reset()
			info_text.Write(headers, text.WriteCellOpts(cell.Bold()))
			for _, item := range message {
				key := strings.Split(item, " ")[0]
				if key == ActiveConns[0] {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGreen))
				} else if colorflipper < 0 {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorGray))
				} else {
					color = text.WriteCellOpts(cell.FgColor(cell.ColorWhite))
				}
				colorflipper *= -1

				info_text.Write(item, color)
			}
		}
	}
}
