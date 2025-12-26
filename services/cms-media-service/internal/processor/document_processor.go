package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DocumentProcessor handles document processing operations
type DocumentProcessor struct {
	outputDir string
}

// NewDocumentProcessor creates a new document processor
func NewDocumentProcessor(outputDir string) *DocumentProcessor {
	return &DocumentProcessor{
		outputDir: outputDir,
	}
}

// ExtractThumbnail extracts first page as thumbnail from PDF
func (dp *DocumentProcessor) ExtractPDFThumbnail(pdfPath, outputPath string) error {
	// Using ImageMagick/convert to extract first page
	args := []string{
		fmt.Sprintf("%s[0]", pdfPath), // [0] means first page
		"-thumbnail", "400x400>",
		"-background", "white",
		"-alpha", "remove",
		"-quality", "85",
		outputPath,
	}
	
	cmd := exec.Command("convert", args...)
	return cmd.Run()
}

// ExtractDocxThumbnail extracts thumbnail from DOCX file
func (dp *DocumentProcessor) ExtractDocxThumbnail(docxPath, outputPath string) error {
	// First convert DOCX to PDF using LibreOffice
	tmpDir := filepath.Join(dp.outputDir, "tmp")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)
	
	// Convert to PDF
	cmd := exec.Command("libreoffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		docxPath,
	)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docx to pdf conversion failed: %w", err)
	}
	
	// Find generated PDF
	baseName := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath))
	pdfPath := filepath.Join(tmpDir, baseName+".pdf")
	
	// Extract thumbnail from PDF
	return dp.ExtractPDFThumbnail(pdfPath, outputPath)
}

// ExtractPPTXThumbnail extracts thumbnail from PPTX file
func (dp *DocumentProcessor) ExtractPPTXThumbnail(pptxPath, outputPath string) error {
	// Similar to DOCX - convert to PDF first then extract thumbnail
	tmpDir := filepath.Join(dp.outputDir, "tmp")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)
	
	// Convert to PDF
	cmd := exec.Command("libreoffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", tmpDir,
		pptxPath,
	)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pptx to pdf conversion failed: %w", err)
	}
	
	// Find generated PDF
	baseName := strings.TrimSuffix(filepath.Base(pptxPath), filepath.Ext(pptxPath))
	pdfPath := filepath.Join(tmpDir, baseName+".pdf")
	
	// Extract thumbnail from PDF
	return dp.ExtractPDFThumbnail(pdfPath, outputPath)
}

// ExtractThumbnailByType extracts thumbnail based on file type
func (dp *DocumentProcessor) ExtractThumbnailByType(filePath, outputPath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	switch ext {
	case ".pdf":
		return dp.ExtractPDFThumbnail(filePath, outputPath)
	case ".docx", ".doc":
		return dp.ExtractDocxThumbnail(filePath, outputPath)
	case ".pptx", ".ppt":
		return dp.ExtractPPTXThumbnail(filePath, outputPath)
	default:
		return fmt.Errorf("unsupported document type: %s", ext)
	}
}

// GetPDFPageCount gets the number of pages in a PDF
func (dp *DocumentProcessor) GetPDFPageCount(pdfPath string) (int, error) {
	cmd := exec.Command("pdfinfo", pdfPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	// Parse output for "Pages:" line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Pages:") {
			var pages int
			_, err := fmt.Sscanf(line, "Pages: %d", &pages)
			if err == nil {
				return pages, nil
			}
		}
	}
	
	return 0, fmt.Errorf("could not determine page count")
}
