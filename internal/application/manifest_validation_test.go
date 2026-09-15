package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote.futrx-cli/internal/project"
)

func TestReadManifestRejectsMalformedAndUnknownFields(t *testing.T) {
	for _, raw := range []string{`{"id":`, `{"id":"demo","unknown":true}`} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "application.json"), []byte(raw), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := project.ReadManifest(dir); err == nil || !strings.Contains(err.Error(), "parse application.json") {
			t.Fatalf("ReadManifest(%q) error = %v", raw, err)
		}
	}
}
