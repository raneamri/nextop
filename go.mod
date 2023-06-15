module github.com/raneamri/gotop

go 1.20

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/mum4k/termdash v0.18.0
)

require (
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell/v2 v2.6.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	golang.org/x/sys v0.9.0 // indirect
	golang.org/x/term v0.9.0 // indirect
	golang.org/x/text v0.10.0 // indirect
)

replace (
	github.com/raneamri/gotop/errors => ./errors
	github.com/raneamri/gotop/io => ./io
	github.com/raneamri/gotop/services => ./services
	github.com/raneamri/gotop/types => ./types
	github.com/raneamri/gotop/ui => ./ui
	github.com/raneamri/gotop/util => ./util
)
