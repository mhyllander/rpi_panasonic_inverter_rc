DEPS=

all: bin/decode

all-rpi: bin/decode-aarch64

bin/decode: cmd/decode/main.go
	go build -o bin/decode cmd/decode/main.go

bin/decode-aarch64: cmd/decode/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/decode-aarch64 cmd/decode/main.go

bin/hello-aarch64: cmd/cgo/main_cgo.go
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/hello-aarch64 -ldflags="--sysroot=/home/mhy/chroot/rpi-bookworm-arm64" cmd/cgo/main_cgo.go
