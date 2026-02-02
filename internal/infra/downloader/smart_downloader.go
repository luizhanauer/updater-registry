package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ArtifactResult contém os dados físicos do arquivo remoto
type ArtifactResult struct {
	Checksum string // SHA256
	Size     int64  // Bytes
}

type SmartDownloader struct {
	client *http.Client
}

func New() *SmartDownloader {
	return &SmartDownloader{
		client: &http.Client{
			Timeout: 5 * time.Minute, // Timeout generoso para downloads grandes
			Transport: &http.Transport{
				DisableKeepAlives: true, // Evita problemas com conexões "penduradas" em CLI
			},
		},
	}
}

// DownloadAndHash baixa o stream para a memória (ou temp file) apenas para calcular o hash
func (d *SmartDownloader) DownloadAndHash(url string) (*ArtifactResult, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Updater-Registry-Bot/1.0 (+https://github.com/seu-repo)")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falha na conexão: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status http inválido: %d", resp.StatusCode)
	}

	// Criamos o Hasher
	hasher := sha256.New()

	// Usamos io.Copy para ler o stream diretamente para o hasher.
	// Isso evita carregar 100MB+ na RAM de uma vez.
	// O retorno 'size' é o número de bytes copiados.
	size, err := io.Copy(hasher, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler stream: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	return &ArtifactResult{
		Checksum: checksum,
		Size:     size,
	}, nil
}
