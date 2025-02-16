package handlers

import (
	"archive/zip"
	"bytes"
	"file-conv/internal/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func MergePDFs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 50MB limit
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pdfmerge-")
	if err != nil {
		http.Error(w, "Error creating temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Process uploaded files
	files, err := utils.ProcessUploadedFiles(r, tempDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Merge PDFs
	mergedPath := filepath.Join(tempDir, "merged.pdf")
	config := model.NewDefaultConfiguration()
	if err := api.MergeCreateFile(files, mergedPath, false, config); err != nil {
		http.Error(w, "Error merging PDFs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read merged file
	mergedBytes, err := os.ReadFile(mergedPath)
	if err != nil {
		http.Error(w, "Error reading merged PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=merged.pdf")
	if _, err := io.Copy(w, bytes.NewReader(mergedBytes)); err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
	}
}

func SplitPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with 50MB limit
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Process uploaded file
	file, header, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "pdfsplit-")
	if err != nil {
		http.Error(w, "Error creating temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Save uploaded file to temp directory
	pdfPath := filepath.Join(tempDir, header.Filename)
	outFile, err := os.Create(pdfPath)
	if err != nil {
		http.Error(w, "Error saving uploaded file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(w, "Error copying file content", http.StatusInternalServerError)
		return
	}

	// Reopen the saved PDF file
	pdfFile, err := os.Open(pdfPath)
	if err != nil {
		http.Error(w, "Error opening saved PDF file", http.StatusInternalServerError)
		return
	}
	defer pdfFile.Close()

	// Get split mode (pages or count)
	mode := r.FormValue("mode")
	config := model.NewDefaultConfiguration()

	switch mode {
	case "pages":
		// Split by specific page numbers
		pageRanges := r.FormValue("pages")
		// Convert pageRanges to a slice of integers
		pages, err := utils.ParsePageRanges(pageRanges)
		if err != nil {
			http.Error(w, "Invalid page ranges", http.StatusBadRequest)
			return
		}
		// Use SplitByPageNrFile for splitting by specific pages
		if err := api.SplitByPageNrFile(pdfPath, tempDir, pages, config); err != nil {
			http.Error(w, "Error splitting PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
	case "count":
		// Split by page count
		pageCountStr := r.FormValue("count")
		pageCount, err := strconv.Atoi(pageCountStr)
		if err != nil || pageCount < 1 {
			http.Error(w, "Invalid page count", http.StatusBadRequest)
			return
		}
		// Use Split with the specified span
		if err := api.Split(pdfFile, tempDir, header.Filename, pageCount, config); err != nil {
			http.Error(w, "Error splitting PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Invalid split mode", http.StatusBadRequest)
		return
	}

	// Create zip archive of split files
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		// Create a zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = info.Name()
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := io.Copy(writer, fileToZip); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Error creating zip archive", http.StatusInternalServerError)
		return
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "Error closing zip writer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=split_pdfs.zip")
	if _, err := io.Copy(w, &buf); err != nil {
		http.Error(w, "Error sending response", http.StatusInternalServerError)
	}
}

func CompressPDFHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form with a 50MB limit
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Retrieve the uploaded PDF file
	file, header, err := r.FormFile("pdf")
	if err != nil {
		http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "pdfcompress-")
	if err != nil {
		http.Error(w, "Error creating temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	// Save the uploaded file to the temporary directory
	inputPath := filepath.Join(tempDir, header.Filename)
	outFile, err := os.Create(inputPath)
	if err != nil {
		http.Error(w, "Error saving uploaded file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(w, "Error copying file content", http.StatusInternalServerError)
		return
	}

	// Define the output path for the compressed PDF
	outputPath := filepath.Join(tempDir, "compressed_"+header.Filename)

	// Compress the PDF using pdfcpu
	config := model.NewDefaultConfiguration()
	if err := api.OptimizeFile(inputPath, outputPath, config); err != nil {
		http.Error(w, "Error compressing PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read the compressed PDF
	compressedFile, err := os.Open(outputPath)
	if err != nil {
		http.Error(w, "Error opening compressed PDF", http.StatusInternalServerError)
		return
	}
	defer compressedFile.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, compressedFile); err != nil {
		http.Error(w, "Error reading compressed PDF", http.StatusInternalServerError)
		return
	}

	// Send the compressed PDF as the response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", "compressed_"+header.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, "Error sending compressed PDF", http.StatusInternalServerError)
	}
}
