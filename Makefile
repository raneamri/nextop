# Build executable in current OS
build:
	go build -o nextop
.PHONY:build

# Build and run
run:
	go build -o nextop && ./nextop
.PHONY:run

# Test & builds
test1:
	cd ./benchmarking &&\
	./suite1.sh && cd ..
	go build -o nextop &&\
	./nextop

test2:
	cd ./benchmarking &&\
	./suite2.sh && cd ..
	go build -o nextop &&\
	./nextop

# Build Linux executable
genlinux:
	CGO_ENABLED=0 go build -o build/nextop-linux_static .
.PHONY:genlinux

# Build Windows executable
genwin: 
	CGO_ENABLED=0 GOOS=windows go build -o build/nextop-win.exe .
.PHONY:genwin

# Build MacOS executable
genmac:
	CGO_ENABLED=0 GOOS=darwin go build -o build/nextop-macos .
.PHONY:genmac

# Build for Debug
debug:
	go build -gcflags="all=-N -l -m" -o nextop
.PHONY:debug

# Build all executables
genall: genmac genwin genlinux