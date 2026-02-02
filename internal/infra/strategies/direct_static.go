package strategies

import (
	"context"
	"time"
	"updater-registry/internal/core/domain"
)

type DirectStaticStrategy struct{}

func NewDirectStaticStrategy() *DirectStaticStrategy {
	return &DirectStaticStrategy{}
}

func (s *DirectStaticStrategy) Name() string { return "direct_static" }

func (s *DirectStaticStrategy) Fetch(ctx context.Context, config map[string]string) (*domain.StrategyResult, error) {
	url := config["url"]

	// Para links estáticos, a versão é a data corrente.
	// O UseCase (Service) detectará que, mesmo a versão sendo "hoje",
	// ele deve verificar o Hash para confirmar se houve update real.
	version := time.Now().Format("2006.01.02")

	return &domain.StrategyResult{
		Version:     version,
		DownloadURL: url,
	}, nil
}
