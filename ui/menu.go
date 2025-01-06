package ui

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"

	_ "github.com/go-sql-driver/mysql"
)

/*
Workload:

	None (static display)
*/

/*
Format:

	widget-1 (left): keybinds, info
	widget-2 (right): logo
*/
func DisplayMenu() {
	t, err := tcell.New()
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	help_table1, _ := text.New()
	help_table1.Write(
		"\n P Processlist\n D InnoDB Dashboard\n M Memory Dashboard\n E Error Log\n L Lock Log\n R Replication\n T Transactions\n C Configs\n ? Help\n ESC Previous Page\n Q Quit",
		text.WriteCellOpts(cell.Bold()))

	help_table2, _ := text.New()
	help_table2.Write(
		"\n TAB Reload page\n -> Cycle to next connection\n <- Cycle to previous connection\n \\ Clear all filters\n / Clear group filters\n = Pause\n + Increase refresh rate by 100ms\n - Decrease refresh rate by 100ms\n _ Export Processlist\n",
		text.WriteCellOpts(cell.Bold()),
	)

	help_table3, _ := text.New()
	help_table3.Write(
		"\n Trying to fix the aspect ratio? Adjust this page's size until the logo on the right is correctly\n in frame. (164x54)\n\n REPO https://github.com/raneamri/nextop\n AUTHOR Imrane AMRI\n LICENSE MIT\n",
		text.WriteCellOpts(cell.Bold()),
	)

	logo_sd, _ := segmentdisplay.New()

	clock_sd, err := segmentdisplay.New()

	go animateLogo(logo_sd)
	go clock(ctx, clock_sd)

	cont, err := container.New(
		t,
		container.ID("menu_screen"),
		container.Border(linestyle.Light),
		container.BorderTitle("NEXTOP (ESC to go back)"),
		container.BorderColor(cell.ColorGray),
		container.FocusedColor(cell.ColorWhite),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Modes"),
						container.PlaceWidget(help_table1),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle("Actions"),
								container.PlaceWidget(help_table2),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle("Other"),
								container.PlaceWidget(help_table3),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.Right(
				container.Border(linestyle.Light),
				container.SplitHorizontal(
					container.Top(),
					container.Bottom(
						container.PlaceWidget(clock_sd),
					),
					container.SplitPercent(15),
				),
			),
			container.SplitPercent(65),
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
		case 'p', 'P':
			State = types.PROCESSLIST
			cancel()
		case 'd', 'D':
			State = types.DB_DASHBOARD
			cancel()
		case 'm', 'M':
			State = types.MEM_DASHBOARD
			cancel()
		case 'e', 'E':
			State = types.ERR_LOG
			cancel()
		case 'l', 'L':
			State = types.LOCK_LOG
			cancel()
		case 'r', 'R':
			State = types.REPLICATION
			cancel()
		case 't', 'T':
			State = types.TRANSACTIONS
			cancel()
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case keyboard.KeyCtrlD:
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
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

func animateLogo(sd *segmentdisplay.SegmentDisplay) {
	var (
		logo string = " NEXTOP"
		out  string
		rndc []string = []string{"$", "%", "&", "*", "@", ":", "/", "?", "<", ">"}
	)

	for _, char := range logo {
		out += string(char)

		for range rndc {
			out += rndc[rand.Intn(len(rndc)-1)]
			chunks := []*segmentdisplay.TextChunk{
				segmentdisplay.NewChunk(out, segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorGray)), segmentdisplay.WriteCellOpts(cell.Bold())),
			}

			sd.Write(chunks)
			out = out[:len(out)-1]
			time.Sleep(25 * time.Millisecond)
			sd.Reset()
		}
	}

	chunks := []*segmentdisplay.TextChunk{
		segmentdisplay.NewChunk(out, segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorWhite)), segmentdisplay.WriteCellOpts(cell.Bold())),
	}
	sd.Write(chunks)
	return
}

func clock(ctx context.Context, sd *segmentdisplay.SegmentDisplay) {

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			nowStr := now.Format("15 04")
			parts := strings.Split(nowStr, " ")

			spacer := " "
			if now.Second()%2 == 0 {
				spacer = ":"
			}
			chunks := []*segmentdisplay.TextChunk{
				segmentdisplay.NewChunk(parts[0], segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorLime)), segmentdisplay.WriteCellOpts(cell.Bold())),
				segmentdisplay.NewChunk(spacer),
				segmentdisplay.NewChunk(parts[1], segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorLime)), segmentdisplay.WriteCellOpts(cell.Bold())),
			}
			if err := sd.Write(chunks); err != nil {
				panic(err)
			}

		case <-ctx.Done():
			return
		}
	}
}
