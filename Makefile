BINARIES=decode paninv_rc
BINARIES_RPI=decode-aarch64 paninv_rc-aarch64

build: $(BINARIES)

build-rpi: $(BINARIES_RPI)

decode: cmd/decode/main.go
	go build -o bin/decode cmd/decode/main.go

paninv_rc: cmd/paninv_rc/main.go
	go build -o bin/paninv_rc cmd/paninv_rc/main.go

decode-aarch64: cmd/decode/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/decode-aarch64 cmd/decode/main.go

paninv_rc-aarch64: cmd/paninv_rc/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/paninv_rc-aarch64 cmd/paninv_rc/main.go

hello-aarch64: cmd/cgo/main_cgo.go
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/hello-aarch64 -ldflags="--sysroot=/home/mhy/chroot/rpi-bookworm-arm64" cmd/cgo/main_cgo.go
