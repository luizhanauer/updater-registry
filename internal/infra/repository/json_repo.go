package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"updater-registry/internal/core/domain"
)

type FileRepository struct {
	baseDir string
}

func NewFileRepository(baseDir string) *FileRepository {
	return &FileRepository{baseDir: baseDir}
}

func (r *FileRepository) Get(id string) (*domain.Package, error) {
	path := filepath.Join(r.baseDir, id+".json")

	// Se arquivo não existe, retornamos nil sem erro (é um pacote novo)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg domain.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("json corrompido em %s: %w", id, err)
	}
	return &pkg, nil
}

func (r *FileRepository) Save(pkg *domain.Package) error {
	path := filepath.Join(r.baseDir, pkg.ID+".json")

	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}

	// Garante que a pasta existe
	if err := os.MkdirAll(r.baseDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (r *FileRepository) ListAll() ([]*domain.Package, error) {
	var packages []*domain.Package

	// Busca todos os arquivos .json na pasta packages/
	files, err := filepath.Glob(filepath.Join(r.baseDir, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Pula arquivos corrompidos ou ilegíveis
		}

		var pkg domain.Package
		if err := json.Unmarshal(data, &pkg); err != nil {
			continue // Pula JSON inválido
		}
		packages = append(packages, &pkg)
	}

	return packages, nil
}
