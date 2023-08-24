package ui

import (
	"context"
	"fmt"
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

*/

/*
Format:
*/
func DisplayThreadAnalysis() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	var (
		lookup  map[string]func() string
		message []string
		fetch   [][]string
	)

	analysis_text, _ := text.New()

	_, err = strconv.Atoi(CurrentThread)
	if err != nil {
		message = append(message, "Non-numeric Thread ID\nESC to return to previous view")
	} else if len(CurrentQuery) == 0 {
		message = append(message, "Invalid or inactive thread")
	} else {
		lookup = GlobalQueryMap[Instances[ActiveConns[0]].DBMS]
		fetch = queries.GetLongQuery(Instances[ActiveConns[0]].Driver, fmt.Sprintf(lookup["thread_analysis"](), CurrentQuery))
	}

	if len(fetch) > 0 {
		message = fetch[0]
	}

	if len(CurrentQuery) > 0 {
		analysis_text.Write("Query: " + CurrentQuery + "\n\n")
	}

	for _, line := range message {
		analysis_text.Write(line + "\n")
	}

	cont, err := container.New(
		t,
		container.ID("thread_analysis"),
		container.Border(linestyle.Light),
		container.BorderTitle("THREAD ANALYSIS (ESC to return)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.PlaceWidget(analysis_text),
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
			CurrentQuery = ""
			State = types.MENU
			cancel()
		case keyboard.KeyTab:
			cancel()
		case keyboard.KeyEsc:
			CurrentQuery = ""
			State = Laststate
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(keyreader), termdash.RedrawInterval(Interval)); err != nil {
		panic(err)
	}
}
