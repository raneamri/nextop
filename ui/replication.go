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
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"
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

	fmt.Printf("Hello world!")

	cont, err := container.New(
		t,
		container.ID("replication"),
		container.Border(linestyle.Light),
		container.BorderTitle("REPLICATION (? for help)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
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
