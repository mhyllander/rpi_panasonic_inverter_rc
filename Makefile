BINARIES=decode pinv_irc
BINARIES_RPI=decode-aarch64 pinv_irc-aarch64

build: $(BINARIES)

build-rpi: $(BINARIES_RPI)

decode: cmd/decode/main.go
	go build -o bin/decode cmd/decode/main.go

pinv_irc: cmd/pinv_irc/main.go
	go build -o bin/pinv_irc cmd/pinv_irc/main.go

decode-aarch64: cmd/decode/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/decode-aarch64 cmd/decode/main.go

pinv_irc-aarch64: cmd/pinv_irc/main.go
	GOOS=linux GOARCH=arm64 go build -o bin/pinv_irc-aarch64 cmd/pinv_irc/main.go

hello-aarch64: cmd/cgo/main_cgo.go
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/hello-aarch64 -ldflags="--sysroot=/home/mhy/chroot/rpi-bookworm-arm64" cmd/cgo/main_cgo.go
