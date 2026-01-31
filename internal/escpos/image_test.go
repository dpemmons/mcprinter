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

func TestEncodeImage_DitheringProducesMixedPixels(t *testing.T) {
	// Create a 48x1 image with a smooth gray gradient
	// With dithering, mid-gray should produce a mix of black and white pixels
	// rather than all-black or all-white
	img := image.NewRGBA(image.Rect(0, 0, 48, 1))
	for x := 0; x < 48; x++ {
		g := uint8(128) // mid-gray
		img.Set(x, 0, color.RGBA{g, g, g, 255})
	}

	data := escpos.EncodeImage(img)

	cmdIdx := bytes.Index(data, []byte{0x1D, 0x76, 0x30, 0x00})
	if cmdIdx == -1 {
		t.Fatal("missing raster command")
	}
	rasterStart := cmdIdx + 8
	// 384 dots wide / 8 = 48 bytes per row, 1 row
	// Count set bits -- dithered mid-gray should have roughly half set
	setBits := 0
	for i := rasterStart; i < len(data); i++ {
		for bit := 0; bit < 8; bit++ {
			if data[i]&(1<<uint(bit)) != 0 {
				setBits++
			}
		}
	}
	totalBits := (len(data) - rasterStart) * 8
	ratio := float64(setBits) / float64(totalBits)
	// Mid-gray dithered should be roughly 30-70% black (not 0% or 100%)
	if ratio < 0.1 || ratio > 0.9 {
		t.Errorf("dithering produced %.0f%% black pixels for mid-gray, expected mixed", ratio*100)
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
