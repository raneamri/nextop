package ui

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/raneamri/gotop/types"
)

func DrawMenu(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	display, err := segmentdisplay.New()
	if err != nil {
		panic(err)
	}
	if err := display.Write([]*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk(fmt.Sprintf("%s", "MENU")),
	}); err != nil {
		panic(err)
	}

	cont, err := container.New(
		t,
		container.ID("main_menu"),
		container.Border(linestyle.Light),
		container.BorderTitle("GOTOP (? for help, ENTER for quickstart)"),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' {
			State = types.QUIT
			cancel()
		} else if k.Key == '?' {
			State = types.HELP
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}
}

func DrawHelp(t *tcell.Terminal) {
	ctx, cancel := context.WithCancel(context.Background())

	cont, err := container.New(
		t,
		container.ID("help_screen"),
		container.Border(linestyle.Light),
		container.BorderTitle("GOTOP/HELP (b to go back)"),
	)
	if err != nil {
		panic(err)
	}

	/*
		Note: add help screen
	*/

	quitter := func(k *terminalapi.Keyboard) {
		switch k.Key {
		case 'b', 'B':
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, cont, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(100*time.Millisecond)); err != nil {
		panic(err)
	}

}
