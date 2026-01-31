# mcprinter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI that sends text and images to a WiFi ESC/POS receipt printer over TCP.

**Architecture:** Cobra CLI resolves positional args into a queue of print items (images then text). An ESC/POS encoder converts items to byte sequences. A TCP transport sends the bytes to the printer.

**Tech Stack:** Go 1.22, Cobra (CLI), godotenv (.env loading), stdlib image packages (image/png, image/jpeg), golang.org/x/image/draw (resizing)

---

### Task 1: Initialize Go Module and Install Dependencies

**Files:**
- Create: `go.mod`
- Create: `go.sum`

**Step 1: Initialize Go module**

Run: `go mod init github.com/dpemmons/mcprinter`

**Step 2: Install dependencies**

Run: `go get github.com/spf13/cobra@latest && go get github.com/joho/godotenv@latest && go get golang.org/x/image@latest`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat: initialize go module with dependencies"
```

---

### Task 2: TCP Transport

**Files:**
- Create: `internal/transport/tcp.go`
- Test: `internal/transport/tcp_test.go`

**Step 1: Write the failing test**

`internal/transport/tcp_test.go`:

```go
package transport_test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/dpemmons/mcprinter/internal/transport"
)

func TestSend_WritesDataToServer(t *testing.T) {
	// Start a TCP listener to simulate a printer
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		data, _ := io.ReadAll(conn)
		received <- data
	}()

	addr := ln.Addr().String()
	host, port, _ := net.SplitHostPort(addr)

	err = transport.Send(host, port, []byte("hello printer"))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case data := <-received:
		if string(data) != "hello printer" {
			t.Errorf("got %q, want %q", string(data), "hello printer")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for data")
	}
}

func TestSend_ErrorOnUnreachableHost(t *testing.T) {
	// Use a port that nothing is listening on
	err := transport.Send("127.0.0.1", "1", []byte("data"))
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/transport/ -v`
Expected: FAIL - package not found

**Step 3: Write minimal implementation**

`internal/transport/tcp.go`:

```go
package transport

import (
	"fmt"
	"net"
	"time"
)

// Send connects to host:port via TCP and writes data.
func Send(host, port string, data []byte) error {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot reach printer at %s: %w", addr, err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed writing to printer: %w", err)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/transport/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/transport/
git commit -m "feat: add TCP transport for sending data to printer"
```

---

### Task 3: ESC/POS Text Encoder

**Files:**
- Create: `internal/escpos/encoder.go`
- Test: `internal/escpos/encoder_test.go`

**Step 1: Write the failing test**

`internal/escpos/encoder_test.go`:

```go
package escpos_test

import (
	"bytes"
	"testing"

	"github.com/dpemmons/mcprinter/internal/escpos"
)

func TestEncodeText(t *testing.T) {
	data := escpos.EncodeText("Hello")

	// Must start with ESC @ (initialize)
	if !bytes.HasPrefix(data, []byte{0x1B, 0x40}) {
		t.Error("missing ESC @ init command")
	}

	// Must contain the text
	if !bytes.Contains(data, []byte("Hello")) {
		t.Error("missing text content")
	}

	// Must end with feed + cut (GS V \x00 = full cut)
	if !bytes.HasSuffix(data, []byte{0x1D, 0x56, 0x00}) {
		t.Error("missing cut command at end")
	}
}

func TestEncodeText_MultipleLines(t *testing.T) {
	data := escpos.EncodeText("Line1\nLine2")

	if !bytes.Contains(data, []byte("Line1\nLine2")) {
		t.Error("missing multi-line text content")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/escpos/ -v`
Expected: FAIL - package not found

**Step 3: Write minimal implementation**

`internal/escpos/encoder.go`:

```go
package escpos

// ESC/POS command bytes
var (
	CmdInit    = []byte{0x1B, 0x40}       // ESC @ - initialize printer
	CmdFeed    = []byte{0x1B, 0x64, 0x04} // ESC d 4 - feed 4 lines
	CmdFullCut = []byte{0x1D, 0x56, 0x00} // GS V 0 - full cut
)

// EncodeText converts a text string into ESC/POS bytes with init and cut.
func EncodeText(text string) []byte {
	var buf []byte
	buf = append(buf, CmdInit...)
	buf = append(buf, []byte(text)...)
	buf = append(buf, '\n')
	buf = append(buf, CmdFeed...)
	buf = append(buf, CmdFullCut...)
	return buf
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/escpos/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/escpos/
git commit -m "feat: add ESC/POS text encoder"
```

---

### Task 4: ESC/POS Image Encoder

**Files:**
- Create: `internal/escpos/image.go`
- Test: `internal/escpos/image_test.go`

**Step 1: Write the failing test**

`internal/escpos/image_test.go`:

```go
package escpos_test

import (
	"bytes"
	"image"
	"image/color"
	"testing"

	"github.com/dpemmons/mcprinter/internal/escpos"
)

func TestEncodeImage_ContainsRasterCommand(t *testing.T) {
	// Create a small 8x2 black image
	img := image.NewRGBA(image.Rect(0, 0, 8, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.Black)
		}
	}

	data := escpos.EncodeImage(img)

	// Must contain GS v 0 (raster bit image command)
	if !bytes.Contains(data, []byte{0x1D, 0x76, 0x30, 0x00}) {
		t.Error("missing GS v 0 raster command")
	}
}

func TestEncodeImage_ResizesToPrinterWidth(t *testing.T) {
	// Create a 768px wide image (2x printer width)
	img := image.NewRGBA(image.Rect(0, 0, 768, 100))

	data := escpos.EncodeImage(img)

	// Should produce data (not panic or return empty)
	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestEncodeImage_WhitePixelsAreZeroBits(t *testing.T) {
	// Create a small 8x1 white image
	img := image.NewRGBA(image.Rect(0, 0, 8, 1))
	for x := 0; x < 8; x++ {
		img.Set(x, 0, color.White)
	}

	data := escpos.EncodeImage(img)

	// The raster data for 8 white pixels = 1 byte of 0x00
	// Command: GS v 0 \x00 + xL xH yL yH + data
	// xL=1 xH=0 (1 byte per row), yL=1 yH=0 (1 row)
	// Find the raster data after the header
	cmdIdx := bytes.Index(data, []byte{0x1D, 0x76, 0x30, 0x00})
	if cmdIdx == -1 {
		t.Fatal("missing raster command")
	}
	// Skip command (4) + params (4) = 8 bytes to get to raster data
	rasterStart := cmdIdx + 8
	if rasterStart >= len(data) {
		t.Fatal("no raster data found")
	}
	if data[rasterStart] != 0x00 {
		t.Errorf("white pixels should be 0x00, got 0x%02X", data[rasterStart])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/escpos/ -v -run Image`
Expected: FAIL - EncodeImage not found

**Step 3: Write minimal implementation**

`internal/escpos/image.go`:

```go
package escpos

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

const PrinterWidth = 384 // dots at 203 DPI for 80mm paper

// EncodeImage converts an image to ESC/POS raster bit image bytes.
// The image is resized to PrinterWidth and converted to 1-bit B&W.
func EncodeImage(img image.Image) []byte {
	// Resize to printer width, maintaining aspect ratio
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	newW := PrinterWidth
	newH := origH * newW / origW

	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Convert to 1-bit raster data
	// Each row is ceil(width/8) bytes, each bit = 1 pixel (1=black, 0=white)
	bytesPerRow := (newW + 7) / 8
	rasterData := make([]byte, bytesPerRow*newH)

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			// Convert to grayscale, threshold at 128
			gray := (r*299 + g*587 + b*114) / 1000
			if gray>>8 < 128 {
				// Black pixel = 1 bit
				byteIdx := y*bytesPerRow + x/8
				bitIdx := uint(7 - x%8)
				rasterData[byteIdx] |= 1 << bitIdx
			}
		}
	}

	// Build GS v 0 command
	// Format: GS v 0 m xL xH yL yH [data]
	// m=0 (normal), x=bytes per row, y=number of rows
	var buf []byte
	buf = append(buf, 0x1D, 0x76, 0x30, 0x00) // GS v 0, m=0
	buf = append(buf, byte(bytesPerRow&0xFF))   // xL
	buf = append(buf, byte(bytesPerRow>>8))     // xH
	buf = append(buf, byte(newH&0xFF))          // yL
	buf = append(buf, byte(newH>>8))            // yH
	buf = append(buf, rasterData...)

	return buf
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./internal/escpos/ -v -run Image`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/escpos/image.go internal/escpos/image_test.go
git commit -m "feat: add ESC/POS raster image encoder"
```

---

### Task 5: Cobra CLI with Smart Arg Resolution

**Files:**
- Create: `cmd/root.go`
- Create: `main.go`
- Create: `.env.example`
- Test: `cmd/root_test.go`

**Step 1: Write the failing test**

`cmd/root_test.go`:

```go
package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dpemmons/mcprinter/cmd"
)

func TestResolveArgs_LiteralText(t *testing.T) {
	items, err := cmd.ResolveArgs([]string{"Hello, world!"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != cmd.ItemText {
		t.Errorf("expected text, got %v", items[0].Type)
	}
	if items[0].Data != "Hello, world!" {
		t.Errorf("got %q", items[0].Data)
	}
}

func TestResolveArgs_TextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "receipt.txt")
	os.WriteFile(path, []byte("Order #1234"), 0644)

	items, err := cmd.ResolveArgs([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != cmd.ItemText {
		t.Fatalf("expected 1 text item, got %d items", len(items))
	}
	if items[0].Data != "Order #1234" {
		t.Errorf("got %q", items[0].Data)
	}
}

func TestResolveArgs_ImageFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "logo.png")
	os.WriteFile(path, []byte("fake png"), 0644)

	items, err := cmd.ResolveArgs([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != cmd.ItemImage {
		t.Fatalf("expected 1 image item, got %+v", items)
	}
	if items[0].Path != path {
		t.Errorf("got path %q", items[0].Path)
	}
}

func TestResolveArgs_ImageThenText(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "logo.png")
	os.WriteFile(imgPath, []byte("fake"), 0644)

	items, err := cmd.ResolveArgs([]string{imgPath, "Thank you!"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Images should sort before text
	if items[0].Type != cmd.ItemImage {
		t.Error("expected image first")
	}
	if items[1].Type != cmd.ItemText {
		t.Error("expected text second")
	}
}

func TestResolveArgs_NoArgs(t *testing.T) {
	_, err := cmd.ResolveArgs([]string{})
	if err == nil {
		t.Fatal("expected error for no args")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./cmd/ -v`
Expected: FAIL - package not found

**Step 3: Write implementation**

`cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/dpemmons/mcprinter/internal/escpos"
	"github.com/dpemmons/mcprinter/internal/transport"

	"image/jpeg"
	"image/png"
)

type ItemType int

const (
	ItemText  ItemType = iota
	ItemImage
)

type PrintItem struct {
	Type ItemType
	Data string // text content (for ItemText)
	Path string // file path (for ItemImage)
}

var (
	flagHost string
	flagPort string
)

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".bmp": true,
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return imageExts[ext]
}

// ResolveArgs takes positional args and resolves them into ordered PrintItems.
// Images are sorted before text.
func ResolveArgs(args []string) ([]PrintItem, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no input provided; pass text, a file path, or pipe via stdin")
	}

	var images []PrintItem
	var texts []PrintItem

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err == nil && !info.IsDir() {
			// It's a file
			if isImageFile(arg) {
				images = append(images, PrintItem{Type: ItemImage, Path: arg})
			} else {
				content, err := os.ReadFile(arg)
				if err != nil {
					return nil, fmt.Errorf("reading %s: %w", arg, err)
				}
				texts = append(texts, PrintItem{Type: ItemText, Data: string(content)})
			}
		} else {
			// Treat as literal text
			texts = append(texts, PrintItem{Type: ItemText, Data: arg})
		}
	}

	// Images first, then text
	items := append(images, texts...)
	return items, nil
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcprinter [flags] [text|file|image]...",
		Short: "Print to a WiFi ESC/POS receipt printer",
		Long:  "Send text and images to a WiFi-connected ESC/POS thermal receipt printer.",
		RunE:  runPrint,
	}

	rootCmd.Flags().StringVar(&flagHost, "host", "", "Printer IP address (overrides PRINTER_HOST)")
	rootCmd.Flags().StringVar(&flagPort, "port", "", "Printer port (overrides PRINTER_PORT)")

	return rootCmd
}

func loadConfig() (host, port string, err error) {
	// Try .env in current dir, then home dir
	godotenv.Load(".env")
	if home, e := os.UserHomeDir(); e == nil {
		godotenv.Load(filepath.Join(home, ".env"))
	}

	host = os.Getenv("PRINTER_HOST")
	port = os.Getenv("PRINTER_PORT")

	if flagHost != "" {
		host = flagHost
	}
	if flagPort != "" {
		port = flagPort
	}
	if port == "" {
		port = "9100"
	}
	if host == "" {
		return "", "", fmt.Errorf("no printer host configured; set PRINTER_HOST in .env or use --host")
	}
	return host, port, nil
}

func runPrint(cmd *cobra.Command, args []string) error {
	// Check stdin
	stdinItems, err := readStdin()
	if err != nil {
		return err
	}

	items, err := ResolveArgs(append(args, stdinItemsToArgs(stdinItems)...))
	if len(args) == 0 && len(stdinItems) == 0 {
		return fmt.Errorf("no input provided; pass text, a file path, or pipe via stdin")
	}
	if err != nil && len(stdinItems) == 0 {
		return err
	}

	// If only stdin, use those items
	if len(args) == 0 {
		items = stdinItems
	}

	host, port, err := loadConfig()
	if err != nil {
		return err
	}

	// Encode all items
	var payload []byte
	payload = append(payload, escpos.CmdInit...)

	for _, item := range items {
		switch item.Type {
		case ItemImage:
			img, err := loadImage(item.Path)
			if err != nil {
				return fmt.Errorf("loading image %s: %w", item.Path, err)
			}
			payload = append(payload, escpos.EncodeImage(img)...)
		case ItemText:
			// Append raw text (init already sent, cut will come at end)
			payload = append(payload, []byte(item.Data)...)
			payload = append(payload, '\n')
		}
	}

	payload = append(payload, escpos.CmdFeed...)
	payload = append(payload, escpos.CmdFullCut...)

	return transport.Send(host, port, payload)
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Decode(f)
	case ".jpg", ".jpeg":
		return jpeg.Decode(f)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}

func readStdin() ([]PrintItem, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, nil
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		// Is a TTY, no piped data
		return nil, nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	return []PrintItem{{Type: ItemText, Data: string(data)}}, nil
}

func stdinItemsToArgs(items []PrintItem) []string {
	// This is a helper so we don't duplicate items - stdin items
	// get handled separately in runPrint
	return nil
}

// Needed for image loading
import "image"

// Ensure stable sort: images before text
func init() {
	sort.SliceStable(nil, func(i, j int) bool { return false })
}
```

Wait — that file has import issues. Let me fix the plan. The `image` import needs to be in the import block, and we don't need the `sort` init. Let me provide a clean version:

`cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/dpemmons/mcprinter/internal/escpos"
	"github.com/dpemmons/mcprinter/internal/transport"
)

// ItemType represents the type of content to print.
type ItemType int

const (
	ItemText  ItemType = iota
	ItemImage
)

// PrintItem holds a resolved piece of content to print.
type PrintItem struct {
	Type ItemType
	Data string // text content (for ItemText)
	Path string // file path (for ItemImage)
}

var (
	flagHost string
	flagPort string
)

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".bmp": true,
}

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return imageExts[ext]
}

// ResolveArgs takes positional args and resolves them into ordered PrintItems.
// Images are sorted before text.
func ResolveArgs(args []string) ([]PrintItem, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no input provided; pass text, a file path, or pipe via stdin")
	}

	var images []PrintItem
	var texts []PrintItem

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err == nil && !info.IsDir() {
			if isImageFile(arg) {
				images = append(images, PrintItem{Type: ItemImage, Path: arg})
			} else {
				content, err := os.ReadFile(arg)
				if err != nil {
					return nil, fmt.Errorf("reading %s: %w", arg, err)
				}
				texts = append(texts, PrintItem{Type: ItemText, Data: string(content)})
			}
		} else {
			texts = append(texts, PrintItem{Type: ItemText, Data: arg})
		}
	}

	return append(images, texts...), nil
}

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcprinter [flags] [text|file|image]...",
		Short: "Print to a WiFi ESC/POS receipt printer",
		Long:  "Send text and images to a WiFi-connected ESC/POS thermal receipt printer.",
		RunE:  runPrint,
	}

	rootCmd.Flags().StringVar(&flagHost, "host", "", "Printer IP address (overrides PRINTER_HOST)")
	rootCmd.Flags().StringVar(&flagPort, "port", "", "Printer port (overrides PRINTER_PORT)")

	return rootCmd
}

func loadConfig() (string, string, error) {
	godotenv.Load(".env")
	if home, err := os.UserHomeDir(); err == nil {
		godotenv.Load(filepath.Join(home, ".env"))
	}

	host := os.Getenv("PRINTER_HOST")
	port := os.Getenv("PRINTER_PORT")

	if flagHost != "" {
		host = flagHost
	}
	if flagPort != "" {
		port = flagPort
	}
	if port == "" {
		port = "9100"
	}
	if host == "" {
		return "", "", fmt.Errorf("no printer host configured; set PRINTER_HOST in .env or use --host")
	}
	return host, port, nil
}

func runPrint(c *cobra.Command, args []string) error {
	var allItems []PrintItem

	// Check stdin for piped data
	if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		if len(data) > 0 {
			allItems = append(allItems, PrintItem{Type: ItemText, Data: string(data)})
		}
	}

	if len(args) > 0 {
		resolved, err := ResolveArgs(args)
		if err != nil {
			return err
		}
		// Merge: images from resolved go first, then stdin text, then resolved text
		var images, texts []PrintItem
		for _, item := range resolved {
			if item.Type == ItemImage {
				images = append(images, item)
			} else {
				texts = append(texts, item)
			}
		}
		allItems = append(images, append(allItems, texts...)...)
	}

	if len(allItems) == 0 {
		return fmt.Errorf("no input provided; pass text, a file path, or pipe via stdin")
	}

	host, port, err := loadConfig()
	if err != nil {
		return err
	}

	var payload []byte
	payload = append(payload, escpos.CmdInit...)

	for _, item := range allItems {
		switch item.Type {
		case ItemImage:
			img, err := loadImageFile(item.Path)
			if err != nil {
				return fmt.Errorf("loading image %s: %w", item.Path, err)
			}
			payload = append(payload, escpos.EncodeImage(img)...)
		case ItemText:
			payload = append(payload, []byte(item.Data)...)
			payload = append(payload, '\n')
		}
	}

	payload = append(payload, escpos.CmdFeed...)
	payload = append(payload, escpos.CmdFullCut...)

	return transport.Send(host, port, payload)
}

func loadImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Decode(f)
	case ".jpg", ".jpeg":
		return jpeg.Decode(f)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
}
```

`main.go`:

```go
package main

import (
	"os"

	"github.com/dpemmons/mcprinter/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
```

`.env.example`:

```
PRINTER_HOST=192.168.1.100
PRINTER_PORT=9100
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./cmd/ -v`
Expected: PASS

**Step 5: Build and verify**

Run: `cd /Users/dpemmons/src/mcprinter && go build -o mcprinter .`
Expected: Binary built successfully

**Step 6: Commit**

```bash
git add main.go cmd/ .env.example
git commit -m "feat: add Cobra CLI with smart arg resolution"
```

---

### Task 6: Integration Smoke Test

**Files:**
- Create: `integration_test.go`

**Step 1: Write the test**

`integration_test.go`:

```go
package main_test

import (
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration_TextPrint(t *testing.T) {
	// Build binary
	dir := t.TempDir()
	bin := filepath.Join(dir, "mcprinter")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = getProjectRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	// Start fake printer
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		data, _ := io.ReadAll(conn)
		received <- data
	}()

	host, port, _ := net.SplitHostPort(ln.Addr().String())

	cmd := exec.Command(bin, "--host", host, "--port", port, "Hello, printer!")
	cmd.Env = append(os.Environ(), "PRINTER_HOST=", "PRINTER_PORT=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcprinter failed: %s\n%s", err, out)
	}

	select {
	case data := <-received:
		if len(data) == 0 {
			t.Fatal("received empty data")
		}
		// Should contain the text somewhere in the ESC/POS stream
		if !containsBytes(data, []byte("Hello, printer!")) {
			t.Error("printed data does not contain expected text")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for print data")
	}
}

func containsBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Assumes test is run from project root
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
```

**Step 2: Run the test**

Run: `cd /Users/dpemmons/src/mcprinter && go test -v -run Integration`
Expected: PASS

**Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration smoke test"
```

---

### Task 7: Final Build and Verify

**Step 1: Run all tests**

Run: `cd /Users/dpemmons/src/mcprinter && go test ./... -v`
Expected: All PASS

**Step 2: Build final binary**

Run: `cd /Users/dpemmons/src/mcprinter && go build -o mcprinter . && ./mcprinter --help`
Expected: Help text displays correctly

**Step 3: Commit any remaining files**

```bash
git add -A
git commit -m "chore: finalize v1 build"
```
