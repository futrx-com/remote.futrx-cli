package application

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var appID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var envKey = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

func Validate(dir string) (Manifest, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return Manifest{}, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, "application.json"))
	if err != nil {
		return Manifest{}, fmt.Errorf("read application.json: %w", err)
	}
	var m Manifest
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return m, fmt.Errorf("parse application.json: %w", err)
	}
	if !appID.MatchString(m.ID) {
		return m, fmt.Errorf("application id %q must use lowercase letters, numbers, and hyphens", m.ID)
	}
	if strings.TrimSpace(m.Name) == "" {
		return m, fmt.Errorf("application.json: name is required")
	}
	if strings.TrimSpace(m.Version) == "" {
		return m, fmt.Errorf("application.json: version is required")
	}
	if len(m.Scopes) == 0 {
		return m, fmt.Errorf("application.json: at least one scope is required")
	}
	for _, scope := range m.Scopes {
		if scope != "global" && scope != "project" {
			return m, fmt.Errorf("application.json: invalid scope %q", scope)
		}
	}
	for _, e := range m.Env {
		if !envKey.MatchString(e.Key) {
			return m, fmt.Errorf("application.json: invalid env key %q", e.Key)
		}
		if e.Generate != "" && e.Generate != "password" {
			return m, fmt.Errorf("application.json: unsupported generator %q", e.Generate)
		}
	}
	install := m.Install
	if install == "" {
		if regular(filepath.Join(dir, "infra", "install.sh")) {
			install = "infra/install.sh"
		}
	}
	if install != "" {
		clean := filepath.ToSlash(filepath.Clean(install))
		if clean != install || !strings.HasPrefix(clean, "infra/") || !regular(filepath.Join(dir, filepath.FromSlash(clean))) {
			return m, fmt.Errorf("application.json: install must name a regular file inside infra/")
		}
	} else if m.Port.Internal != 0 || m.Port.DefaultExternal != 0 || m.Service != "" || m.Healthcheck.Command != "" {
		return m, fmt.Errorf("port, service, and healthcheck require infra/install.sh")
	}
	if m.Port.Internal < 0 || m.Port.Internal > 65535 || m.Port.DefaultExternal < 0 || m.Port.DefaultExternal > 65535 {
		return m, fmt.Errorf("application.json: ports must be between 1 and 65535")
	}
	if m.Port.Internal == 0 && (m.Port.DefaultExternal != 0 || m.Healthcheck.Command != "") {
		return m, fmt.Errorf("defaultExternal and healthcheck require port.internal")
	}
	if m.Port.Protocol != "" && m.Port.Protocol != "tcp" && m.Port.Protocol != "udp" {
		return m, fmt.Errorf("application.json: protocol must be tcp or udp")
	}
	hasCapability := install != "" || isDir(filepath.Join(dir, "ui")) || isDir(filepath.Join(dir, "backend")) || isDir(filepath.Join(dir, "skills"))
	if !hasCapability {
		return m, fmt.Errorf("application has no infra, UI, backend, or skills capability")
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			rel, _ := filepath.Rel(dir, path)
			return fmt.Errorf("symlinks are not supported: %s", rel)
		}
		return nil
	})
	if err != nil {
		return m, err
	}
	return m, nil
}

func regular(path string) bool { i, e := os.Stat(path); return e == nil && i.Mode().IsRegular() }
func isDir(path string) bool   { i, e := os.Stat(path); return e == nil && i.IsDir() }
