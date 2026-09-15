package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
)

func Scaffold(parent, name string) (string, error) {
	id := slug(name)
	if !domain.ApplicationID.MatchString(id) {
		return "", fmt.Errorf("%q does not produce a valid application id", name)
	}
	dir := filepath.Join(parent, id)
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("%s already exists", dir)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	for rel, body := range scaffoldFiles(id, displayName(name)) {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", err
		}
		mode := os.FileMode(0644)
		if rel == "infra/install.sh" {
			mode = 0755
		}
		if err := os.WriteFile(path, []byte(body), mode); err != nil {
			return "", err
		}
	}
	abs, _ := filepath.Abs(dir)
	return abs, nil
}

func slug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if b.Len() > 0 && !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
func displayName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
