package utils

import (
	"image"
	"image/color"
)

func DetectBackgroundColor(img image.Image) color.Color {
	bounds := img.Bounds()
	colorCount := make(map[color.RGBA]int)

	// Sample edges to detect background color
	sampleWidth := 10
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Min.X+sampleWidth; x++ {
			CountColor(colorCount, img.At(x, y))
		}
		for x := bounds.Max.X - sampleWidth; x < bounds.Max.X; x++ {
			CountColor(colorCount, img.At(x, y))
		}
	}

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Min.Y+sampleWidth; y++ {
			CountColor(colorCount, img.At(x, y))
		}
		for y := bounds.Max.Y - sampleWidth; y < bounds.Max.Y; y++ {
			CountColor(colorCount, img.At(x, y))
		}
	}

	// Find the most common color
	var maxColor color.RGBA
	maxCount := 0
	for c, count := range colorCount {
		if count > maxCount {
			maxColor = c
			maxCount = count
		}
	}

	return maxColor
}
func CountColor(colorCount map[color.RGBA]int, c color.Color) {
	r, g, b, a := c.RGBA()
	key := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
	colorCount[key]++
}

func IsColorMatch(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	// Allow a tolerance to account for slight variations
	tolerance := uint32(5000)
	return AbsDiff(r1, r2) < tolerance &&
		AbsDiff(g1, g2) < tolerance &&
		AbsDiff(b1, b2) < tolerance
}

func AbsDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}
