package strategies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"updater-registry/internal/core/domain"
)

type GithubStrategy struct {
	client *http.Client
}

func NewGithubStrategy() *GithubStrategy {
	return &GithubStrategy{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *GithubStrategy) Name() string { return "github_release" }

func (s *GithubStrategy) Fetch(ctx context.Context, config map[string]string) (*domain.StrategyResult, error) {
	repo := config["repo"]
	assetFilter := config["asset_filter"]

	if repo == "" {
		return nil, fmt.Errorf("config 'repo' é obrigatória")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	// Importante para evitar Rate Limit no GitHub Actions
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var rel struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}

	// Normaliza versão (remove 'v' prefixo)
	version := strings.TrimPrefix(rel.TagName, "v")

	// Procura o asset correto
	for _, asset := range rel.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(assetFilter)) {
			return &domain.StrategyResult{
				Version:     version,
				DownloadURL: asset.BrowserDownloadURL,
			}, nil
		}
	}

	return nil, fmt.Errorf("nenhum asset contendo '%s' encontrado na release %s", assetFilter, version)
}
