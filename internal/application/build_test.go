package application

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeVersion(t *testing.T) {
	for input, want := range map[string]string{"1.2.3": "1.2.3", " release 1 / beta ": "release-1---beta", "---": ""} {
		if got := safeVersion(input); got != want {
			t.Errorf("safeVersion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuildDefaultOutputAndResult(t *testing.T) {
	parent := t.TempDir()
	dir, err := Scaffold(parent, "demo")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Build(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Path != filepath.Join(parent, "demo-0.1.0.zip") || result.ID != "demo" || result.Version != "0.1.0" || len(result.SHA256) != 64 || result.Files == 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestBuildRejectsOutputInsideApplication(t *testing.T) {
	dir, err := Scaffold(t.TempDir(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Build(dir, filepath.Join(dir, "dist", "demo.zip"))
	if err == nil || !strings.Contains(err.Error(), "outside the application") {
		t.Fatalf("error = %v", err)
	}
}

func TestArchiveIsSortedAndNormalized(t *testing.T) {
	raw, err := buildApplicationArchive([]sourceFile{{name: "z", data: []byte("z")}, {name: "a", data: []byte("a")}})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if zr.File[0].Name != "a" || zr.File[1].Name != "z" {
		t.Fatalf("order = %s, %s", zr.File[0].Name, zr.File[1].Name)
	}
	for _, file := range zr.File {
		if !file.Modified.Equal(archiveEpoch) || file.Mode().Perm() != 0644 {
			t.Errorf("metadata for %s not normalized", file.Name)
		}
	}
}

func TestInfraPayloadSelectionAndContents(t *testing.T) {
	files := []sourceFile{{name: "infra/install.sh", data: []byte("install")}, {name: "infra/config", data: []byte("config")}, {name: "backend/main.go", data: []byte("go")}}
	raw, ok, err := infraPayload(files)
	if err != nil || !ok {
		t.Fatalf("infraPayload = ok %v, err %v", ok, err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "infra/config" {
		t.Fatalf("payload entry = %q", header.Name)
	}
	if _, err := tr.Next(); err == nil {
		t.Fatal("unexpected second payload entry")
	}
	if _, ok, err := infraPayload([]sourceFile{{name: "infra/install.sh"}}); err != nil || ok {
		t.Fatalf("empty payload = ok %v, err %v", ok, err)
	}
}

func TestCollectSourceFilesExcludesDevelopmentArtifacts(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{"keep.txt": "yes", ".git/config": "no", "dist/out": "no", ".gitignore": "no", "old.zip": "no", "infra/package.sh": "no", "infra/payload.tar.gz": "no"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	files, err := collectSourceFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].name != "keep.txt" {
		t.Fatalf("files = %+v", files)
	}
}

func TestCollectSourceFilesRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := collectSourceFiles(root); err == nil || !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("error = %v", err)
	}
}
