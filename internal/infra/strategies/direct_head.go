package strategies

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"updater-registry/internal/core/domain"
)

type DirectHeadStrategy struct {
	client *http.Client
}

func NewDirectHeadStrategy() *DirectHeadStrategy {
	return &DirectHeadStrategy{
		client: &http.Client{
			Timeout:       10 * time.Second,
			CheckRedirect: nil, // Segue redirects automaticamente
		},
	}
}

func (s *DirectHeadStrategy) Name() string { return "direct_url_head" }

func (s *DirectHeadStrategy) Fetch(ctx context.Context, config map[string]string) (*domain.StrategyResult, error) {
	startURL := config["url"]
	regexPattern := config["regex"] // Pode ser vazio para Static (Chrome)

	req, _ := http.NewRequestWithContext(ctx, "HEAD", startURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d na url %s", resp.StatusCode, startURL)
	}

	finalURL := resp.Request.URL.String()

	// 1. Tenta extrair versão via Regex (se configurado)
	version := ""
	if regexPattern != "" {
		re := regexp.MustCompile(regexPattern)
		matches := re.FindStringSubmatch(finalURL)
		if len(matches) >= 2 {
			version = matches[1]
		}
	}

	// 2. Extrai Metadados dos Headers (Otimização do VS Code e Chrome)
	remoteChecksum := resp.Header.Get("x-sha256") // VS Code envia isso!

	// Tenta Content-Length (Discord/Chrome)
	var remoteSize int64
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		remoteSize, _ = strconv.ParseInt(cl, 10, 64)
	}
	// Fallback para x-goog-stored-content-length (alguns buckets do google)
	if remoteSize == 0 {
		if cl := resp.Header.Get("x-goog-stored-content-length"); cl != "" {
			remoteSize, _ = strconv.ParseInt(cl, 10, 64)
		}
	}

	return &domain.StrategyResult{
		Version:        version, // Pode ser vazio (ex: Chrome), o Service lidará com isso
		DownloadURL:    finalURL,
		RemoteChecksum: remoteChecksum,
		RemoteSize:     remoteSize,
	}, nil
}
