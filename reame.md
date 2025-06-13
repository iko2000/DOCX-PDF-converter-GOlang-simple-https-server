# DOCX to PDF Converter

A simple Go HTTP server that converts DOCX files to PDF using the UniOffice library.

## Features

- Upload DOCX files via web interface
- Convert to PDF format
- Download converted files
- Clean API endpoints
- Automatic file cleanup

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/docx-to-pdf-converter.git
cd docx-to-pdf-converter
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run main.go
```

4. Open http://localhost:8080 in your browser

## Usage

### Web Interface
- Go to http://localhost:8080
- Select a DOCX file
- Click "Convert to PDF"
- Download the result

### API
```bash
# Upload and convert
curl -X POST -F "docx=@document.docx" http://localhost:8080/convert

# Download converted file
curl -O http://localhost:8080/download/filename.pdf
```

## Requirements

- Go 1.21+
- UniOffice library

**Note**: UniOffice requires a license for commercial use. For personal/open source projects, you can get a free license at https://cloud.unidoc.io

## License

MIT License - feel free to use this in your own projects.

## Contributing

Pull requests welcome! This is a simple project meant to help developers who need DOCX to PDF conversion.