package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
)

func ReadManifest(dir string) (domain.Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "application.json"))
	if err != nil {
		return domain.Manifest{}, fmt.Errorf("read application.json: %w", err)
	}
	var manifest domain.Manifest
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, fmt.Errorf("parse application.json: %w", err)
	}
	return manifest, nil
}
