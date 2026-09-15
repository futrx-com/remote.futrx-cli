package project

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
)

func ResolveInstall(dir string, manifest domain.Manifest) (string, error) {
	install := manifest.Install
	if install == "" && isRegular(filepath.Join(dir, "infra", "install.sh")) {
		install = "infra/install.sh"
	}
	if install != "" {
		clean := filepath.ToSlash(filepath.Clean(install))
		if clean != install || !strings.HasPrefix(clean, "infra/") || !isRegular(filepath.Join(dir, filepath.FromSlash(clean))) {
			return "", fmt.Errorf("application.json: install must name a regular file inside infra/")
		}
	} else if manifest.Port.Internal != 0 || manifest.Port.DefaultExternal != 0 || manifest.Service != "" || manifest.Healthcheck.Command != "" {
		return "", fmt.Errorf("port, service, and healthcheck require infra/install.sh")
	}
	return install, nil
}

func HasCapability(dir, install string) bool {
	return install != "" || isDirectory(filepath.Join(dir, "ui")) || isDirectory(filepath.Join(dir, "backend")) || isDirectory(filepath.Join(dir, "skills"))
}

func ValidateTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			relative, _ := filepath.Rel(dir, path)
			return fmt.Errorf("symlinks are not supported: %s", relative)
		}
		return nil
	})
}

func isRegular(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
func isDirectory(path string) bool { info, err := os.Stat(path); return err == nil && info.IsDir() }
