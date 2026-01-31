package cmd

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/spf13/cobra"

	"github.com/dpemmons/mcprinter/internal/escpos"
	"github.com/dpemmons/mcprinter/internal/transport"
)

func NewCalibrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calibrate",
		Short: "Print a calibration page to determine printer width in dots",
		RunE:  runCalibrate,
	}
}

func runCalibrate(c *cobra.Command, args []string) error {
	host, port, err := loadConfig()
	if err != nil {
		return err
	}

	img := generateCalibrationImage()

	var payload []byte
	payload = append(payload, escpos.CmdInit...)
	payload = append(payload, escpos.EncodeImage(img)...)
	payload = append(payload, escpos.CmdFeed...)
	payload = append(payload, escpos.CmdFullCut...)

	fmt.Println("Printing calibration page...")
	fmt.Println("The last fully visible dashed line indicates your printer width.")
	fmt.Println("Set PRINTER_WIDTH in your .env to that value.")

	return transport.Send(host, port, payload)
}

func generateCalibrationImage() image.Image {
	// Test widths covering common printer resolutions
	widths := []int{192, 256, 288, 320, 384, 432, 480, 512, 546, 576}

	maxWidth := widths[len(widths)-1]
	rowHeight := 30 // pixels per calibration row
	dashHeight := 4
	labelHeight := rowHeight - dashHeight
	_ = labelHeight
	imgHeight := len(widths) * rowHeight

	img := image.NewRGBA(image.Rect(0, 0, maxWidth, imgHeight))
	// Fill white
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	for i, w := range widths {
		yBase := i * rowHeight

		// Draw dashed line at this width
		dashLen := 8
		gapLen := 4
		for x := 0; x < w; x++ {
			pos := x % (dashLen + gapLen)
			if pos < dashLen {
				for dy := 0; dy < dashHeight; dy++ {
					img.Set(x, yBase+dy, color.Black)
				}
			}
		}

		// Draw the width number as pixel digits below the dashes
		label := fmt.Sprintf("%d", w)
		drawPixelText(img, 2, yBase+dashHeight+2, label)
	}

	return img
}

// drawPixelText renders a string using a tiny 3x5 pixel font.
func drawPixelText(img *image.RGBA, x, y int, s string) {
	for _, ch := range s {
		glyph, ok := font3x5[ch]
		if !ok {
			x += 4
			continue
		}
		for row := 0; row < 5; row++ {
			for col := 0; col < 3; col++ {
				if glyph[row]&(1<<(2-col)) != 0 {
					// Draw 2x scaled for readability
					for dy := 0; dy < 2; dy++ {
						for dx := 0; dx < 2; dx++ {
							img.Set(x+col*2+dx, y+row*2+dy, color.Black)
						}
					}
				}
			}
		}
		x += 8 // 3*2 pixels + 2 spacing
	}
}

// font3x5 is a minimal 3x5 bitmap font for digits. Each row is 3 bits wide.
var font3x5 = map[rune][5]byte{
	'0': {0b111, 0b101, 0b101, 0b101, 0b111},
	'1': {0b010, 0b110, 0b010, 0b010, 0b111},
	'2': {0b111, 0b001, 0b111, 0b100, 0b111},
	'3': {0b111, 0b001, 0b111, 0b001, 0b111},
	'4': {0b101, 0b101, 0b111, 0b001, 0b001},
	'5': {0b111, 0b100, 0b111, 0b001, 0b111},
	'6': {0b111, 0b100, 0b111, 0b101, 0b111},
	'7': {0b111, 0b001, 0b001, 0b001, 0b001},
	'8': {0b111, 0b101, 0b111, 0b101, 0b111},
	'9': {0b111, 0b101, 0b111, 0b001, 0b111},
}
