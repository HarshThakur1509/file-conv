package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	files, err := processUploadedFiles(r, tempDir)
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

func processUploadedFiles(r *http.Request, tempDir string) ([]string, error) {
	var files []string

	// Get all files from the "pdfs" form field
	pdfFiles := r.MultipartForm.File["pdfs"]
	if len(pdfFiles) < 2 {
		return nil, fmt.Errorf("at least two PDF files are required")
	}

	for _, fileHeader := range pdfFiles {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("error opening uploaded file: %v", err)
		}
		defer file.Close()

		tempPath := filepath.Join(tempDir, filepath.Base(fileHeader.Filename))
		outFile, err := os.Create(tempPath)
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %v", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, file); err != nil {
			return nil, fmt.Errorf("error saving uploaded file: %v", err)
		}

		files = append(files, tempPath)
	}

	return files, nil
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
		pages, err := parsePageRanges(pageRanges)
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

// parsePageRanges converts a comma-separated string of page numbers into a slice of integers
func parsePageRanges(pageRanges string) ([]int, error) {
	var pages []int
	ranges := strings.Split(pageRanges, ",")
	for _, r := range ranges {
		page, err := strconv.Atoi(strings.TrimSpace(r))
		if err != nil {
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}
