package domain

import "time"

// Package representa a entidade persistida em packages/*.json
type Package struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Category       string    `json:"category"` // text-editor, developer, etc.
	IconURL        string    `json:"icon_url"`
	PackageName    string    `json:"package_name"` // ex: google-chrome-stable
	InstallType    string    `json:"install_type"` // ex: deb, AppImage
	CurrentRelease *Release  `json:"current_release,omitempty"`
	LastCheckedAt  time.Time `json:"last_checked_at"`
}

// Release é o objeto de valor que muda com frequência
type Release struct {
	Version     string `json:"version"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"` // SHA256 obrigatório
	Size        int64  `json:"size"`     // Bytes
}

// ChecksumMatch verifica se o hash local bate com o remoto (se tivermos essa info)
func (p *Package) IsUpToDate(remoteHash string) bool {
	if p.CurrentRelease == nil {
		return false
	}
	return p.CurrentRelease.Checksum == remoteHash
}

type StrategyResult struct {
	// Version: Se a estratégia conseguiu descobrir via URL/API (ex: github, regex).
	// Se vier vazio, o Service sabe que precisa baixar o arquivo para descobrir.
	Version string

	DownloadURL string

	// Metadados Opcionais (Pre-Check)
	// Se a estratégia pegar um header x-sha256 ou content-length, preenche aqui.
	RemoteChecksum string
	RemoteSize     int64
}
