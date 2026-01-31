package cmd

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
		Use:   "mcprint [flags] [text|file|image]...",
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

	if w := os.Getenv("PRINTER_WIDTH"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			escpos.PrinterWidth = n
		}
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
