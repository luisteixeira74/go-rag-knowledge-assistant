package chunker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// DocumentReader é responsável por ler arquivos físicos (.md, .txt, .pdf) e extrair texto simples
type DocumentReader struct{}

func NewDocumentReader() *DocumentReader {
	return &DocumentReader{}
}

// ReadContent identifica a extensão do arquivo e extrai seu conteúdo em texto plano
func (r *DocumentReader) ReadContent(filePath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".md", ".txt":
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("falha ao ler arquivo %s: %w", filePath, err)
		}
		return string(content), nil

	case ".pdf":
		return r.readPDF(filePath)

	default:
		return "", fmt.Errorf("extensão de arquivo '%s' não é suportada", ext)
	}
}

// readPDF extrai todo o texto plano de um arquivo PDF
func (r *DocumentReader) readPDF(filePath string) (string, error) {
	f, pdfReader, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("falha ao abrir arquivo PDF (%s): %w", filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	plainTextReader, err := pdfReader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("falha ao extrair texto do PDF (%s): %w", filePath, err)
	}

	_, err = buf.ReadFrom(plainTextReader)
	if err != nil {
		return "", fmt.Errorf("falha ao ler buffer de texto do PDF: %w", err)
	}

	return buf.String(), nil
}