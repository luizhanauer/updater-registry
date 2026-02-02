package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"updater-registry/internal/core/domain"
	"updater-registry/internal/core/ports"
	"updater-registry/internal/infra/parser"
)

type SourceConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	IconURL     string            `json:"icon_url"`
	PackageName string            `json:"package_name"`
	InstallType string            `json:"install_type"`
	Strategy    string            `json:"strategy"`
	Config      map[string]string `json:"config"`
}

type UpdaterService struct {
	repo         ports.PackageRepository
	stratFactory ports.StrategyFactory
	debExtractor *parser.DebExtractor // Injeção da ferramenta de parse
}

func NewUpdaterService(r ports.PackageRepository, f ports.StrategyFactory) *UpdaterService {
	return &UpdaterService{
		repo:         r,
		stratFactory: f,
		debExtractor: parser.NewDebExtractor(),
	}
}

func (s *UpdaterService) Process(ctx context.Context, src SourceConfig) error {
	log.Printf("🔹 [%s] Checando...", src.Name)

	// 1. Carrega Estado Anterior
	pkg, _ := s.repo.Get(src.ID)
	if pkg == nil {
		pkg = &domain.Package{
			ID:          src.ID,
			Name:        src.Name,
			Description: src.Description,
			Category:    src.Category,
			IconURL:     src.IconURL,
			PackageName: src.PackageName,
			InstallType: src.InstallType,
		}
	}

	// 2. Executa Estratégia
	strategy, err := s.stratFactory.GetStrategy(src.Strategy)
	if err != nil {
		return err
	}

	res, err := strategy.Fetch(ctx, src.Config)
	if err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	// 3. Verificações de "Cache" (Evitar Download)

	// A) Hash Remoto Confiável (VS Code Header x-sha256)
	if res.RemoteChecksum != "" && pkg.CurrentRelease != nil {
		if res.RemoteChecksum == pkg.CurrentRelease.Checksum {
			log.Printf("   ✅ [%s] Hash remoto (header) coincide. Nada a fazer.", src.ID)
			pkg.LastCheckedAt = time.Now()
			return s.repo.Save(pkg)
		}
		log.Printf("   ⚠️ [%s] Hash remoto mudou! (%s -> %s)", src.ID, pkg.CurrentRelease.Checksum[:8], res.RemoteChecksum[:8])
	}

	// B) Comparação de Versão (Se a estratégia descobriu, ex: GitHub/Discord Regex)
	if res.Version != "" && pkg.CurrentRelease != nil {
		if res.Version == pkg.CurrentRelease.Version {
			log.Printf("   ✅ [%s] Versão inalterada (%s).", src.ID, res.Version)
			pkg.LastCheckedAt = time.Now()
			return s.repo.Save(pkg)
		}
	}

	// C) Comparação de Tamanho (Fallback para Static/Chrome)
	// Se não temos versão nem hash confiável, olhamos o tamanho.
	if res.Version == "" && res.RemoteChecksum == "" && pkg.CurrentRelease != nil {
		if res.RemoteSize > 0 && res.RemoteSize == pkg.CurrentRelease.Size {
			log.Printf("   ✅ [%s] Tamanho do arquivo estático idêntico (%d bytes). Mantendo.", src.ID, res.RemoteSize)
			pkg.LastCheckedAt = time.Now()
			return s.repo.Save(pkg)
		}
	}

	// 4. Download Real (Se chegou aqui, precisa atualizar ou validar)
	log.Printf("   ⬇️ [%s] Baixando para inspeção...", src.ID)

	tmpFile, err := os.CreateTemp("", "pkg-*."+src.InstallType)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	size, checksum, err := downloadToFile(res.DownloadURL, tmpFile)
	if err != nil {
		return err
	}

	// Verifica se baixou o que esperava (Double check)
	if res.RemoteChecksum != "" && checksum != res.RemoteChecksum {
		log.Printf("   ⚠️ [%s] Hash calculado difere do header! (Man-in-the-middle ou header errado?)", src.ID)
		// Aqui decidimos se confiamos no arquivo baixado ou abortamos.
		// Geralmente confiamos no arquivo baixado.
	}

	// 5. Extração de Metadados (Versão Real)
	finalVersion := res.Version

	// Se a versão veio vazia (Chrome) OU se queremos garantir a verdade via binário (.deb)
	if finalVersion == "" || src.InstallType == "deb" {
		realVersion, err := s.debExtractor.GetVersion(tmpFile.Name())
		if err == nil && realVersion != "" {
			if finalVersion != "" && finalVersion != realVersion {
				log.Printf("   ℹ️ [%s] Versão corrigida via .deb: %s -> %s", src.ID, finalVersion, realVersion)
			}
			finalVersion = realVersion
		} else {
			log.Printf("   ❌ [%s] Falha ao ler versão do .deb: %v", src.ID, err)
			// Se falhar e não tinhamos versão anterior, é crítico.
			if finalVersion == "" {
				return fmt.Errorf("impossível determinar versão do pacote")
			}
		}
	}

	// 6. Atualiza e Salva
	pkg.CurrentRelease = &domain.Release{
		Version:     finalVersion,
		DownloadURL: res.DownloadURL,
		Checksum:    checksum,
		Size:        size,
	}
	pkg.LastCheckedAt = time.Now()

	log.Printf("   💾 [%s] Salvo: v%s", src.ID, finalVersion)
	return s.repo.Save(pkg)
}

func downloadToFile(url string, f *os.File) (int64, string, error) {
	// ... (mesma implementação anterior com io.MultiWriter)
	resp, err := http.Get(url)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	hasher := sha256.New()
	writer := io.MultiWriter(f, hasher)
	size, err := io.Copy(writer, resp.Body)
	return size, hex.EncodeToString(hasher.Sum(nil)), err
}
