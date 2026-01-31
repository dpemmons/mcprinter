# mcprinter Design

A Go CLI utility for printing to WiFi-based ESC/POS receipt printers.

## Overview

Single-binary Go CLI that sends ESC/POS commands over raw TCP to a WiFi receipt printer. Targets generic ESC/POS over TCP (port 9100), 80mm paper width (384 dots at 203 DPI).

## CLI Usage

```bash
# Print text directly
mcprinter "Hello, world!"

# Print from stdin
echo "Order #1234" | mcprinter

# Print a text file
mcprinter receipt.txt

# Print an image
mcprinter logo.png

# Print image on top, text below
mcprinter logo.png "Thanks for your order!"

# Print image on top, text file below
mcprinter logo.png receipt.txt

# Override connection from .env
mcprinter --host 192.168.1.50 --port 9100 "Hello"
```

## Arg Resolution

For each positional arg (left to right):
- Check if it's an existing file
  - Image file (.png, .jpg, .jpeg, .bmp): queue as image
  - Text file: read contents, queue as text
  - Otherwise: treat as literal text string
- If stdin is not a TTY, read stdin as text
- Print order: all images first, then all text, then auto-cut

`--host` and `--port` flags override .env values. If neither provides a host, exit with a helpful error.

## Configuration

`.env` file loaded from current directory or home directory:

```
PRINTER_HOST=192.168.1.100
PRINTER_PORT=9100
```

## Architecture

```
Input (args/stdin/file) -> Smart resolver -> ESC/POS encoder -> TCP -> Printer
```

### Components

- **CLI layer** (Cobra): Parses args, reads .env, resolves what to print
- **ESC/POS encoder**: Converts text and images into ESC/POS byte sequences
- **Raster engine**: Converts PNG/JPEG to 1-bit raster bitmap for ESC/POS
- **Transport**: Opens TCP connection, sends bytes, closes

## ESC/POS Encoding

### Text
- Initialize printer (ESC @)
- Send raw text bytes with newlines (LF)
- After all content: feed lines, send cut command (GS V)

### Image
- Load image, resize to fit 384 dots wide
- Convert to 1-bit black & white using threshold
- Encode as ESC/POS raster bit image (GS v 0)
- Send raster data row by row

## Transport

- Raw TCP connection
- 5 second connect timeout, 30 second write timeout
- Send all encoded bytes, then close
- Clear error messages on connection failure

## Project Structure

```
mcprinter/
├── main.go              # Entry point
├── cmd/
│   └── root.go          # Cobra root command, arg resolution
├── internal/
│   ├── escpos/
│   │   ├── encoder.go   # Text -> ESC/POS bytes
│   │   └── image.go     # Image -> raster ESC/POS bytes
│   └── transport/
│       └── tcp.go       # TCP connection handling
├── .env.example         # Example config
├── go.mod
└── go.sum
```

## Future Considerations (not in v1)

- Bold, underline, text alignment
- Barcode/QR code printing
- 58mm paper width support
- Configurable DPI
