package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/futrx-com/remote.futrx-cli/internal/domain"
)

func TestReadManifest(t *testing.T) {
	dir := t.TempDir()
	raw := `{"id":"demo","name":"Demo","version":"1","scopes":["project"]}`
	if err := os.WriteFile(filepath.Join(dir, "application.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	manifest, err := ReadManifest(dir)
	if err != nil || manifest.ID != "demo" {
		t.Fatalf("manifest=%+v err=%v", manifest, err)
	}
}

func TestReadManifestMissing(t *testing.T) {
	if _, err := ReadManifest(t.TempDir()); err == nil || !strings.Contains(err.Error(), "read application.json") {
		t.Fatalf("error=%v", err)
	}
}

func TestResolveInstall(t *testing.T) {
	dir := t.TempDir()
	infra := filepath.Join(dir, "infra")
	if err := os.Mkdir(infra, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(infra, "install.sh"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	install, err := ResolveInstall(dir, domain.Manifest{})
	if err != nil || install != "infra/install.sh" {
		t.Fatalf("install=%q err=%v", install, err)
	}
	for _, path := range []string{"../install.sh", "infra/missing.sh", "infra/../install.sh"} {
		if _, err := ResolveInstall(dir, domain.Manifest{Install: path}); err == nil {
			t.Errorf("accepted %q", path)
		}
	}
}

func TestResolveInstallRequiredByRuntime(t *testing.T) {
	for _, manifest := range []domain.Manifest{{Port: domain.Port{Internal: 80}}, {Service: "demo"}, {Healthcheck: domain.Healthcheck{Command: "ok"}}} {
		if _, err := ResolveInstall(t.TempDir(), manifest); err == nil || !strings.Contains(err.Error(), "require infra/install.sh") {
			t.Errorf("manifest=%+v err=%v", manifest, err)
		}
	}
}

func TestHasCapability(t *testing.T) {
	dir := t.TempDir()
	if HasCapability(dir, "") {
		t.Fatal("empty project has capability")
	}
	for _, capability := range []string{"ui", "backend", "skills"} {
		if err := os.Mkdir(filepath.Join(dir, capability), 0755); err != nil {
			t.Fatal(err)
		}
		if !HasCapability(dir, "") {
			t.Errorf("%s directory not detected", capability)
		}
		if err := os.Remove(filepath.Join(dir, capability)); err != nil {
			t.Fatal(err)
		}
	}
	if !HasCapability(dir, "infra/install.sh") {
		t.Fatal("install not detected")
	}
}

func TestValidateTreeRejectsNestedSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", filepath.Join(dir, "nested", "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := ValidateTree(dir); err == nil || !strings.Contains(err.Error(), "nested/link") {
		t.Fatalf("error=%v", err)
	}
}
