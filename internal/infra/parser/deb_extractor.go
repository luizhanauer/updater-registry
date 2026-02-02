package parser

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DebExtractor encapsula a lógica de ler arquivos .deb
type DebExtractor struct{}

func NewDebExtractor() *DebExtractor {
	return &DebExtractor{}
}

// GetVersion executa 'dpkg-deb -f <arquivo> Version'
func (d *DebExtractor) GetVersion(filePath string) (string, error) {
	cmd := exec.Command("dpkg-deb", "--field", filePath, "Version")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("falha ao ler metadados do deb: %w", err)
	}

	// Limpa quebras de linha (ex: "1.0.0\n")
	return strings.TrimSpace(out.String()), nil
}
