# Go Cross-Compilation

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

### Cgo

[Cgo](https://pkg.go.dev/cmd/cgo#hdr-Using_cgo_with_the_go_command) provides a way to link and call C code from a Go program. To use this, you need to install a C cross-compiler on your development system.

Depending on the project, you may also need to have the header and library files from the target system available locally to be able to compile and link correctly. (As it turned out, this was not needed for this project, but see below for a mini how-to.)

For this project it turned out that I needed to use Cgo, but only because I am using [Gorm](https://gorm.io) for database access. To cross-compile it was necessary to enable Cgo and use the C cross-compiler.

#### Installing a C cross-compiler for Raspberry Pi Zero 2 W

```
apt install crossbuild-essential-arm64
```

#### Cross-compiling with Cgo

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc go build -o prog-arm64 main.go
```

## Inspiration

https://dh1tw.de/2019/12/cross-compiling-golang-cgo-projects/

https://gcc.gnu.org/onlinedocs/gcc/Directory-Options.html

https://dev.to/metal3d/understand-how-to-use-c-libraries-in-go-with-cgo-3dbn

https://jensd.be/1126/linux/cross-compiling-for-arm-or-aarch64-on-debian-or-ubuntu

### Creating a Raspberry Pi OS root file system

(This is for reference only, not needed in this project.)

Adapted from [Setting Up a Cross-Compilation Environment](https://earthly.dev/blog/cross-compiling-raspberry-pi/).

#### Collect some info on the Raspberry Pi Zero 2 W

    dpkg --print-architecture
        arm64
    cat /etc/os-release
        PRETTY_NAME="Debian GNU/Linux 12 (bookworm)"
        NAME="Debian GNU/Linux"
        VERSION_ID="12"
        VERSION="12 (bookworm)"
        VERSION_CODENAME=bookworm
        ID=debian
        HOME_URL="https://www.debian.org/"
        SUPPORT_URL="https://www.debian.org/support"
        BUG_REPORT_URL="https://bugs.debian.org/"
    uname -m
        aarch64
    cat /etc/apt/sources.list
        deb http://deb.debian.org/debian bookworm main contrib non-free non-free-firmware
        deb http://deb.debian.org/debian-security/ bookworm-security main contrib non-free non-free-firmware
        deb http://deb.debian.org/debian bookworm-updates main contrib non-free non-free-firmware

#### On Ubuntu development system (e.g WSL2 on Windows)

See <https://ftp-master.debian.org/keys.html> for PGP keys.

    curl -sL https://ftp-master.debian.org/keys/release-12.asc | gpg --import -
    gpg --export F8D2585B8783D481 > $HOME/bookworm-archive-keyring.gpg

    cat > $HOME/rpi.sources <<EOF
    deb http://deb.debian.org/debian bookworm main contrib non-free non-free-firmware
    deb http://deb.debian.org/debian-security/ bookworm-security main contrib non-free non-free-firmware
    deb http://deb.debian.org/debian bookworm-updates main contrib non-free non-free-firmware
    EOF

    cat > $HOME/.mk-sbuild.rc <<EOF
    SOURCE_CHROOTS_DIR="$HOME/chroots"
    DEBOOTSTRAP_KEYRING="$HOME/bookworm-archive-keyring.gpg"
    TEMPLATE_SOURCES="$HOME/rpi.sources"
    SKIP_UPDATES="1"
    SKIP_PROPOSED="1"
    SKIP_SECURITY="1"
    EATMYDATA="1"
    EOF

    export ARCH=arm64
    export RELEASE=bookworm
    mk-sbuild --arch=$ARCH $RELEASE --debootstrap-mirror=http://deb.debian.org/debian --name=rpi-$RELEASE

`mk-sbuild` needs to run as root.
