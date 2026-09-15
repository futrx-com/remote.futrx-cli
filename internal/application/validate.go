package application

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Validate(dir string) (Manifest, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return Manifest{}, err
	}
	manifest, err := readManifest(dir)
	if err != nil {
		return manifest, err
	}
	if err := validateManifestMetadata(manifest); err != nil {
		return manifest, err
	}
	install, err := resolveInstall(dir, manifest)
	if err != nil {
		return manifest, err
	}
	if err := validateManifestRuntime(manifest); err != nil {
		return manifest, err
	}
	if !hasCapability(dir, install) {
		return manifest, fmt.Errorf("application has no infra, UI, backend, or skills capability")
	}
	if err := validateApplicationTree(dir); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func readManifest(dir string) (Manifest, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "application.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read application.json: %w", err)
	}
	var manifest Manifest
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&manifest); err != nil {
		return manifest, fmt.Errorf("parse application.json: %w", err)
	}
	return manifest, nil
}
