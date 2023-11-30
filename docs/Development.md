## Go

## Cross-compiling for the RaspBerry Pi Zero 2 W

### Basic cross-compilation using only GoLang

Cross-compiling using GoLang tools is simple. Set two environment variables to target the Raspberry Pi. For the Raspberry Pi Zero 2:

```bash
export GOOS=linux GOARCH=arm64
go build -o prog-arm64 main.go
```

or

```bash
GOOS=linux GOARCH=arm64 go build -o prog-arm64 main.go
```

### Cross-compilation with cgo

[Cgo](https://pkg.go.dev/cmd/cgo#hdr-Using_cgo_with_the_go_command) provides a way to link and call C code from a Go program. To use this, you need to install a C cross-compiler on your development system. You may also need to have access to the header and library files from the target system to be able to compile and link correctly.

#### Installing a C cross-compiler for Raspberry Pi Zero 2 W

#### Setting up a cross-compilation target environment

## Inspiration

https://dh1tw.de/2019/12/cross-compiling-golang-cgo-projects/
https://gcc.gnu.org/onlinedocs/gcc/Directory-Options.html

https://dev.to/metal3d/understand-how-to-use-c-libraries-in-go-with-cgo-3dbn

https://jensd.be/1126/linux/cross-compiling-for-arm-or-aarch64-on-debian-or-ubuntu

Setting Up a Cross-Compilation Environment
https://earthly.dev/blog/cross-compiling-raspberry-pi/
