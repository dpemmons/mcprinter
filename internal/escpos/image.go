package escpos

import (
	"image"

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
