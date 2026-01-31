package escpos

import (
	"image"

	"golang.org/x/image/draw"
)

var PrinterWidth = 576 // dots, configurable via PRINTER_WIDTH env var

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

	// Floyd-Steinberg dithering to 1-bit
	// Build grayscale float buffer
	gray := make([][]float64, newH)
	for y := 0; y < newH; y++ {
		gray[y] = make([]float64, newW)
		for x := 0; x < newW; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			// Luminance in 0-255 range
			gray[y][x] = float64(r*299+g*587+b*114) / 1000.0 / 256.0
		}
	}

	// Apply Floyd-Steinberg dithering
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			old := gray[y][x]
			var newVal float64
			if old < 128 {
				newVal = 0
			} else {
				newVal = 255
			}
			gray[y][x] = newVal
			err := old - newVal
			if x+1 < newW {
				gray[y][x+1] += err * 7.0 / 16.0
			}
			if y+1 < newH {
				if x-1 >= 0 {
					gray[y+1][x-1] += err * 3.0 / 16.0
				}
				gray[y+1][x] += err * 5.0 / 16.0
				if x+1 < newW {
					gray[y+1][x+1] += err * 1.0 / 16.0
				}
			}
		}
	}

	// Convert to 1-bit raster data
	// Each row is ceil(width/8) bytes, each bit = 1 pixel (1=black, 0=white)
	bytesPerRow := (newW + 7) / 8
	rasterData := make([]byte, bytesPerRow*newH)

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			if gray[y][x] < 128 {
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
