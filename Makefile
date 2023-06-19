# Build executable in current OS
build:
	go build -o gotop
.PHONY:build

# Build and run (quickstart)
run:
	go build -o gotop && ./gotop
.PHONY:run

# Build Linux executable
genlinux:
	CGO_ENABLED=0 go build -o build/gotop-linux_static .
.PHONY:genlinux

# Build Windows executable
genwin: 
	CGO_ENABLED=0 GOOS=windows go build -o build/gotopo-win.exe .
.PHONY:genwin

# Build MacOS executable
genmac:
	CGO_ENABLED=0 GOOS=darwin go build -o build/gotop-macos .
.PHONY:genmac

# Build for Debug
debug:
	go build -gcflags="all=-N -l -m" -o gotop
.PHONY:debug

# Build all executables
genall: genmac genwin genlinux