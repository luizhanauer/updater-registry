package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"updater-registry/internal/core/services"
	"updater-registry/internal/infra/repository"
	"updater-registry/internal/infra/strategies"
)

func main() {
	log.Println("🚀 Iniciando Updater Registry...")

	// 1. Configuração de Caminhos
	cwd, _ := os.Getwd()
	packagesDir := filepath.Join(cwd, "packages")
	apiDir := filepath.Join(cwd, "api")
	sourcesFile := filepath.Join(cwd, "apps.source.json")

	// 2. Inicialização de Dependências
	repo := repository.NewFileRepository(packagesDir)

	// A Factory agora tem o método GetStrategy que a interface pede
	stratFactory := strategies.NewFactory()

	// O Updater espera a interface ports.StrategyFactory
	updater := services.NewUpdaterService(repo, stratFactory)

	// O Exporter agora existe (criado no passo 5 acima)
	exporter := services.NewCatalogExporter(apiDir)

	// 3. Carregar Fontes
	sourcesRaw, err := os.ReadFile(sourcesFile)
	if err != nil {
		log.Fatalf("Erro ao ler apps.source.json: %v", err)
	}
	var sources []services.SourceConfig
	if err := json.Unmarshal(sourcesRaw, &sources); err != nil {
		log.Fatalf("JSON inválido: %v", err)
	}

	// 4. Processamento
	ctx := context.Background()
	errorsCount := 0

	for _, src := range sources {
		if err := updater.Process(ctx, src); err != nil {
			log.Printf("❌ Erro em %s: %v", src.ID, err)
			errorsCount++
		}
	}

	// 5. Geração do Catálogo
	log.Println("📦 Gerando catálogo unificado...")

	// Agora ListAll existe no repo
	allPkgs, err := repo.ListAll()
	if err != nil {
		log.Fatalf("Erro ao listar pacotes: %v", err)
	}

	if err := exporter.Export(allPkgs); err != nil {
		log.Fatalf("Erro ao exportar catálogo: %v", err)
	}

	if errorsCount > 0 {
		log.Printf("⚠️ Finalizado com %d erros.", errorsCount)
		os.Exit(1)
	} else {
		log.Println("✨ Sucesso total.")
	}
}
