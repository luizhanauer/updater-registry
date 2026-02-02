package strategies

import (
	"fmt"
	"updater-registry/internal/core/ports"
)

type Factory struct {
	strategies map[string]ports.UpdateStrategy
}

func NewFactory() *Factory {
	return &Factory{
		strategies: map[string]ports.UpdateStrategy{
			"github_release":  NewGithubStrategy(),
			"direct_url_head": NewDirectHeadStrategy(),
			"direct_static":   NewDirectStaticStrategy(),
		},
	}
}

func (f *Factory) GetStrategy(strategyName string) (ports.UpdateStrategy, error) {
	if strategy, exists := f.strategies[strategyName]; exists {
		return strategy, nil
	}
	return nil, fmt.Errorf("estratégia desconhecida: %s", strategyName)
}
