package ports

import (
	"context"
	"updater-registry/internal/core/domain"
)

// PackageRepository lida com a persistência dos arquivos json individuais
type PackageRepository interface {
	Get(id string) (*domain.Package, error)
	Save(pkg *domain.Package) error
	ListAll() ([]*domain.Package, error)
}

// CatalogExporter lida com a criação do arquivo api/catalog.json
type CatalogExporter interface {
	Export(packages []*domain.Package) error
}

// Strategy define como buscar informações de uma fonte específica
type UpdateStrategy interface {
	Fetch(ctx context.Context, config map[string]string) (*domain.StrategyResult, error)
	Name() string
}

// StrategyFactory seleciona a estratégia correta baseada na string (ex: "github_release")
type StrategyFactory interface {
	GetStrategy(name string) (UpdateStrategy, error)
}

// Downloader abstrai a complexidade de baixar e hashear
type Downloader interface {
	DownloadAndHash(url string) (checksum string, size int64, err error)
}
