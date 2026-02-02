package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
	"updater-registry/internal/core/domain"
)

type CatalogExporter struct {
	outputDir string
}

func NewCatalogExporter(outputDir string) *CatalogExporter {
	return &CatalogExporter{outputDir: outputDir}
}

type CatalogOutput struct {
	LastUpdated time.Time                  `json:"last_updated"`
	Apps        map[string]*domain.Package `json:"apps"`
}

func (e *CatalogExporter) Export(packages []*domain.Package) error {
	// Cria a estrutura final (mapa por ID para fácil acesso)
	catalog := CatalogOutput{
		LastUpdated: time.Now().UTC(),
		Apps:        make(map[string]*domain.Package),
	}

	for _, pkg := range packages {
		catalog.Apps[pkg.ID] = pkg
	}

	// Garante que a pasta api/ existe
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return err
	}

	// Serializa
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}

	// Salva catalog.json
	return os.WriteFile(filepath.Join(e.outputDir, "catalog.json"), data, 0644)
}
