# mcprinter

A command line utility for printing to WiFi-based ESC/POS receipt printers.

Sends text and images over TCP to thermal receipt printers. Supports smart argument resolution -- pass text strings, text files, or images in any order and mcprinter figures out what to do.

## Install

```bash
make
```

This builds the `mcprinter` binary in the project directory.

## Configuration

Copy `.env.example` to `.env` and set your printer's IP:

```bash
cp .env.example .env
```

```
PRINTER_HOST=192.168.1.100
PRINTER_PORT=9100

# Printer width in dots. Common values for 80mm paper:
#   384 = 203 DPI
#   512 = 256 DPI
#   576 = 300 DPI
PRINTER_WIDTH=576
```

Flags `--host` and `--port` override `.env` values.

## Usage

```bash
# Print text
./mcprinter "Hello, world!"

# Print from stdin
echo "Order #1234" | ./mcprinter

# Print a text file
./mcprinter receipt.txt

# Print an image (resized to full paper width, dithered to B&W)
./mcprinter logo.png

# Print image on top, text below
./mcprinter logo.png "Thanks for your order!"

# Print image on top, text file below
./mcprinter logo.png receipt.txt

# Override connection
./mcprinter --host 192.168.1.50 --port 9100 "Hello"
```

### Calibrate

If you're not sure what `PRINTER_WIDTH` to use, print a calibration page:

```bash
./mcprinter calibrate
```

The last fully visible dashed line tells you the correct width in dots.

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
