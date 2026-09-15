package application

import (
	"fmt"
	"path/filepath"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
	"github.com/futrx-com/remote.futrx-cli/internal/project"
)

func Validate(dir string) (Manifest, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return Manifest{}, err
	}
	manifest, err := project.ReadManifest(dir)
	if err != nil {
		return manifest, err
	}
	if err := domain.ValidateMetadata(manifest); err != nil {
		return manifest, err
	}
	install, err := project.ResolveInstall(dir, manifest)
	if err != nil {
		return manifest, err
	}
	if err := domain.ValidateRuntime(manifest); err != nil {
		return manifest, err
	}
	if !project.HasCapability(dir, install) {
		return manifest, fmt.Errorf("application has no infra, UI, backend, or skills capability")
	}
	if err := project.ValidateTree(dir); err != nil {
		return manifest, err
	}
	return manifest, nil
}
