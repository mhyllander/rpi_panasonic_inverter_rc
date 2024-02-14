BINARIES := bin/decode bin/paninv_rc bin/paninv_controller
BINARIES_RPI := arm64/decode arm64/paninv_rc arm64/paninv_controller

all: build build-rpi

build: $(subst /,-,$(BINARIES))

build-rpi: $(subst /,-,$(BINARIES_RPI))

bin-decode:
	go build -o bin/decode cmd/decode/main.go

bin-paninv_rc:
	go build -o bin/paninv_rc cmd/paninv_rc/main.go

bin-paninv_controller:
	go build -o bin/paninv_controller cmd/paninv_controller/main.go

arm64-decode:
	GOOS=linux GOARCH=arm64 go build -o arm64/decode cmd/decode/main.go

arm64-paninv_rc:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o arm64/paninv_rc cmd/paninv_rc/main.go

arm64-paninv_controller:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o arm64/paninv_controller cmd/paninv_controller/main.go

#hello-arm64:
#	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o bin/hello-arm64 -ldflags="--sysroot=/home/mhy/chroot/rpi-bookworm-arm64" cmd/cgo/main_cgo.go

test:
	go test ./codec ./utils

deploy: test build-rpi
	ssh piir 'sudo systemctl stop paninv_controller.service; [ -d bin ] || mkdir bin; [ -d paninv ] && rm -rf paninv/web || mkdir paninv'
	scp $(BINARIES_RPI) piir:bin/
	scp -r web piir:paninv/
	ssh piir sudo systemctl start paninv_controller.service

deploy_jobs:
	scp jobs.json piir:
	ssh piir 'PANINV_DB=paninv/paninv.db bin/paninv_controller -load-jobs jobs.json; sudo systemctl restart paninv_controller.service'

clean:
	rm -f bin/* arm64/*
