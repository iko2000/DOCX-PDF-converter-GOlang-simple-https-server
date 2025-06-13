package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/document/convert"
)

const (
	maxFileSize = 10 << 20 // 10MB
	uploadDir   = "./uploads"
	outputDir   = "./output"
)

func main() {
	// Creates necessary directories!!!!!!!!!!!!!!!!
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Routes to handle all of the incoming exist requests !!!
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/convert", convertHandler)
	http.HandleFunc("/download/", downloadHandler)

	// Starts the server. I have choosen default port 8080. 
	
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("Upload DOCX files to: http://localhost:8080/convert")
	log.Fatal(http.ListenAndServe(port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>DOCX to PDF Converter</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .upload-area { border: 2px dashed #ccc; padding: 40px; text-align: center; margin: 20px 0; }
        button { background: #007cba; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <h1>DOCX to PDF Converter</h1>
    <form action="/convert" method="post" enctype="multipart/form-data">
        <div class="upload-area">
            <input type="file" name="docx" accept=".docx" required>
            <p>Select a DOCX file to convert to PDF</p>
        </div>
        <button type="submit">Convert to PDF</button>
    </form>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Get uploaded file -- UI ENSURES THAT THIS HAPPENS SMOOTH (I hope you have better UI skills:)
	file, header, err := r.FormFile("docx")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".docx") {
		http.Error(w, "Only DOCX files are allowed", http.StatusBadRequest)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102_150405")
	baseFilename := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	docxFilename := fmt.Sprintf("%s_%s.docx", baseFilename, timestamp)
	pdfFilename := fmt.Sprintf("%s_%s.pdf", baseFilename, timestamp)

	// Save uploaded file
	docxPath := filepath.Join(uploadDir, docxFilename)
	dst, err := os.Create(docxPath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Convert DOCX to PDF
	pdfPath := filepath.Join(outputDir, pdfFilename)
	err = convertDocxToPdf(docxPath, pdfPath)
	if err != nil {
		log.Printf("Conversion error: %v", err)
		http.Error(w, fmt.Sprintf("Error converting file: %v", err), http.StatusInternalServerError)
		return
	}

	// Clean up uploaded file
	os.Remove(docxPath)

	// Return success response with download link
	response := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Conversion Complete</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; padding: 15px; border-radius: 5px; }
        a { color: #007cba; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>Conversion Complete!</h1>
    <div class="success">
        <p>Your DOCX file has been successfully converted to PDF.</p>
        <p><a href="/download/%s" download>Download PDF</a></p>
    </div>
    <p><a href="/">Convert another file</a></p>
</body>
</html>`, pdfFilename)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(response))
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	// Security check - prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filepath := filepath.Join(outputDir, filename)
	
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/pdf")

	// Serve file
	http.ServeFile(w, r, filepath)

	// Clean up file after download (optional)
	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(filepath)
	}()
}

func convertDocxToPdf(docxPath, pdfPath string) error {
	// Open the DOCX document
	doc, err := document.Open(docxPath)
	if err != nil {
		return fmt.Errorf("failed to open DOCX file: %w", err)
	}
	defer doc.Close()

	// Convert to PDF - returns a creator.Creator object
	pdfCreator := convert.ConvertToPdf(doc)
	
	// Write the PDF to file
	err = pdfCreator.WriteToFile(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to write PDF file: %w", err)
	}

	return nil
}