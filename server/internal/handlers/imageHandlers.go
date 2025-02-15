package handlers

import (
	"bytes"
	"file-conv/internal/utils"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"github.com/nfnt/resize"
)

func ConvertJPGToPNG(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode JPG", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode PNG", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=converted.png")
	_, _ = io.Copy(w, &buf)
}

func ConvertPNGToJPG(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode PNG", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		http.Error(w, "Failed to encode JPG", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "attachment; filename=converted.jpg")
	_, _ = io.Copy(w, &buf)
}

func CompressImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	qualityStr := r.FormValue("quality")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 1 || quality > 100 {
		quality = 50 // default value
	}

	var buf bytes.Buffer
	if format == "jpeg" {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			http.Error(w, "Failed to compress image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=compressed.jpg")
	} else if format == "png" {
		if err := png.Encode(&buf, img); err != nil {
			http.Error(w, "Failed to compress image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename=compressed.png")
	}

	_, _ = io.Copy(w, &buf)
}

func ResizeImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	// Parse width and height from form values
	widthStr := r.FormValue("width")
	heightStr := r.FormValue("height")

	width, err := strconv.Atoi(widthStr)
	if err != nil || width < 0 {
		width = 0
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 0 {
		height = 0
	}

	if width == 0 && height == 0 {
		http.Error(w, "At least one of width or height must be a positive integer", http.StatusBadRequest)
		return
	}

	// Resize image using Lanczos3 algorithm
	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	var buf bytes.Buffer

	switch format {
	case "jpeg":
		if err := jpeg.Encode(&buf, resizedImg, nil); err != nil {
			http.Error(w, "Failed to encode resized image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=resized.jpg")
	case "png":
		if err := png.Encode(&buf, resizedImg); err != nil {
			http.Error(w, "Failed to encode resized image", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", "attachment; filename=resized.png")
	default:
		http.Error(w, "Unsupported image format", http.StatusBadRequest)
		return
	}

	_, _ = io.Copy(w, &buf)
}

func ConvertToPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Decode image without format check
	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	// Create new RGB image to remove alpha channel
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Create buffer instead of temporary file
	var imgBuf bytes.Buffer
	if err := jpeg.Encode(&imgBuf, rgba, &jpeg.Options{Quality: 90}); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	// Add image from buffer
	pdf.RegisterImageOptionsReader("converted.jpg",
		gofpdf.ImageOptions{ImageType: "JPG"},
		bytes.NewReader(imgBuf.Bytes()),
	)
	pdf.Image("converted.jpg", 10, 10, 190, 0, false, "", 0, "")

	var pdfBuf bytes.Buffer
	if err := pdf.Output(&pdfBuf); err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=converted.pdf")
	_, _ = io.Copy(w, &pdfBuf)
}

func BackgroundTransparent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Detect background color by sampling edges
	backgroundColor := utils.DetectBackgroundColor(img)

	// Iterate over each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			currentColor := img.At(x, y)
			if utils.IsColorMatch(currentColor, backgroundColor) {
				// Set transparent
				rgba.Set(x, y, color.Transparent)
			} else {
				// Copy original color
				rgba.Set(x, y, currentColor)
			}
		}
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "attachment; filename=transparent.png")
	png.Encode(w, rgba)
}
