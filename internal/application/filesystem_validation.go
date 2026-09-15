package application

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func resolveInstall(dir string, manifest Manifest) (string, error) {
	install := manifest.Install
	if install == "" && regular(filepath.Join(dir, "infra", "install.sh")) {
		install = "infra/install.sh"
	}
	if install != "" {
		clean := filepath.ToSlash(filepath.Clean(install))
		if clean != install || !strings.HasPrefix(clean, "infra/") || !regular(filepath.Join(dir, filepath.FromSlash(clean))) {
			return "", fmt.Errorf("application.json: install must name a regular file inside infra/")
		}
	} else if manifest.Port.Internal != 0 || manifest.Port.DefaultExternal != 0 || manifest.Service != "" || manifest.Healthcheck.Command != "" {
		return "", fmt.Errorf("port, service, and healthcheck require infra/install.sh")
	}
	return install, nil
}

func hasCapability(dir, install string) bool {
	return install != "" || isDir(filepath.Join(dir, "ui")) || isDir(filepath.Join(dir, "backend")) || isDir(filepath.Join(dir, "skills"))
}

func validateApplicationTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			rel, _ := filepath.Rel(dir, path)
			return fmt.Errorf("symlinks are not supported: %s", rel)
		}
		return nil
	})
}

func regular(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
