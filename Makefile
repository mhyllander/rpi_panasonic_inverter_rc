BINARIES=decode paninv_rc paninv_controller
BINARIES_RPI=decode-arm64 paninv_rc-arm64 paninv_controller-arm64

all: build build-rpi

build: $(BINARIES)

build-rpi: $(BINARIES_RPI)

decode: cmd/decode/main.go
	go build -o bin/decode cmd/decode/main.go

paninv_rc: cmd/paninv_rc/main.go
	go build -o bin/paninv_rc cmd/paninv_rc/main.go

paninv_controller: cmd/paninv_controller/main.go
	go build -o bin/paninv_controller cmd/paninv_controller/main.go

decode-arm64: cmd/decode/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/decode-arm64 cmd/decode/main.go

paninv_rc-arm64: cmd/paninv_rc/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/paninv_rc-arm64 cmd/paninv_rc/main.go

paninv_controller-arm64: cmd/paninv_controller/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/paninv_controller-arm64 cmd/paninv_controller/main.go

clean:
	rm -f bin/*

#hello-arm64: cmd/cgo/main_cgo.go
#	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/hello-arm64 -ldflags="--sysroot=/home/mhy/chroot/rpi-bookworm-arm64" cmd/cgo/main_cgo.go

deploy: build-rpi
	scp bin/decode-arm64 piir:decode
	scp bin/paninv_rc-arm64 piir:paninv_rc
	scp bin/paninv_controller-arm64 piir:paninv_controller