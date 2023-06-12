module github.com/raneamri/gotop

go 1.20

require (
	github.com/alexeyco/simpletable v1.0.0
	github.com/go-sql-driver/mysql v1.7.1
)

require (
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
)

require (
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/term v0.8.0
)

replace (
	github.com/raneamri/gotop/errors => ./errors
	github.com/raneamri/gotop/io => ./io
	github.com/raneamri/gotop/services => ./services
	github.com/raneamri/gotop/types => ./types
	github.com/raneamri/gotop/ui => ./ui
	github.com/raneamri/gotop/util => ./util
)
