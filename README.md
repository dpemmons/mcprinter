# McPrinter

A command line utility for printing to WiFi-based ESC/POS receipt printers.

Sends text and images over TCP to thermal receipt printers. Supports smart argument resolution -- pass text strings, text files, or images in any order and mcprinter figures out what to do.

## Install

```bash
make
```

This builds the `mcprint` binary in the project directory.

To install to `~/.local/bin`:

```bash
make install
```

## Configuration

`make install` copies `.env.example` to `~/.config/mcprint/config.env` on first run. Edit it with your printer's IP:

```
PRINTER_HOST=192.168.1.100
PRINTER_PORT=9100

# Printer width in dots. Common values for 80mm paper:
#   384 = 203 DPI
#   512 = 256 DPI
#   576 = 300 DPI
PRINTER_WIDTH=576
```

A local `.env` in the working directory is also loaded and takes precedence. Flags `--host` and `--port` override both.

## Usage

```bash
# Print text
./mcprint "Hello, world!"

# Print from stdin
echo "Order #1234" | ./mcprint

# Print a text file
./mcprint soul.txt

# Print an image (resized to full paper width, dithered to B&W)
./mcprint logo.png

# Print image on top, text below
./mcprint logo.png "Thanks for your order!"

# Print image on top, text file below
./mcprint logo.png soul.txt

# Override connection
./mcprint --host 192.168.1.50 --port 9100 "Hello"
```

## Argument Resolution

Arguments are resolved left to right:

- If an arg is an existing file with an image extension (`.png`, `.jpg`, `.jpeg`, `.bmp`), it's queued as an image
- If an arg is an existing file with any other extension, its contents are read as text
- Otherwise, the arg is treated as a literal text string
- If stdin is piped, it's read as text

Print order: images first, then text, then auto-cut.

## Supported Printers

Targets any ESC/POS compatible WiFi receipt printer that accepts raw TCP connections (typically port 9100). Tested with:

- Volcora 80mm Thermal Receipt Printer
