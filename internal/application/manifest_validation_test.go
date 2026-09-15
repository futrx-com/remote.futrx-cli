package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
	"github.com/futrx-com/remote.futrx-cli/internal/project"
)

func validManifest() Manifest {
	return Manifest{ID: "demo", Name: "Demo", Version: "1.0.0", Scopes: []string{"project"}}
}

func TestValidateManifestMetadata(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Manifest)
		want string
	}{
		{"invalid id", func(m *Manifest) { m.ID = "Bad_ID" }, `application id "Bad_ID"`},
		{"missing name", func(m *Manifest) { m.Name = "  " }, "name is required"},
		{"missing version", func(m *Manifest) { m.Version = "" }, "version is required"},
		{"missing scopes", func(m *Manifest) { m.Scopes = nil }, "at least one scope"},
		{"invalid scope", func(m *Manifest) { m.Scopes = []string{"team"} }, `invalid scope "team"`},
		{"invalid env key", func(m *Manifest) { m.Env = []Env{{Key: "lower"}} }, `invalid env key "lower"`},
		{"invalid generator", func(m *Manifest) { m.Env = []Env{{Key: "TOKEN", Generate: "uuid"}} }, `unsupported generator "uuid"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifest()
			tt.edit(&manifest)
			err := domain.ValidateMetadata(manifest)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
	if err := domain.ValidateMetadata(validManifest()); err != nil {
		t.Fatalf("valid manifest: %v", err)
	}
}

func TestValidateManifestRuntime(t *testing.T) {
	tests := []struct {
		name   string
		port   Port
		health string
		want   string
	}{
		{"negative internal", Port{Internal: -1}, "", "ports must be"},
		{"oversized external", Port{Internal: 1, DefaultExternal: 65536}, "", "ports must be"},
		{"external without internal", Port{DefaultExternal: 80}, "", "require port.internal"},
		{"healthcheck without internal", Port{}, "curl localhost", "require port.internal"},
		{"invalid protocol", Port{Internal: 80, Protocol: "http"}, "", "protocol must be"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := validManifest()
			manifest.Port = tt.port
			manifest.Healthcheck.Command = tt.health
			err := domain.ValidateRuntime(manifest)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

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
