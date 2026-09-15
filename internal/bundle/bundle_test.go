package bundle

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestArchiveIsSortedAndNormalized(t *testing.T) {
	raw, err := Archive([]File{{Name: "z", Data: []byte("z")}, {Name: "a", Data: []byte("a")}})
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
		if !file.Modified.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) || file.Mode().Perm() != 0644 {
			t.Errorf("metadata for %s not normalized", file.Name)
		}
	}
}

func TestInfraPayloadSelectionAndContents(t *testing.T) {
	files := []File{{Name: "infra/install.sh", Data: []byte("install")}, {Name: "infra/config", Data: []byte("config")}, {Name: "backend/main.go", Data: []byte("go")}}
	raw, ok, err := InfraPayload(files)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(gz)
	header, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "infra/config" {
		t.Fatalf("entry=%q", header.Name)
	}
	if _, err := reader.Next(); err == nil {
		t.Fatal("unexpected second entry")
	}
	if _, ok, err := InfraPayload([]File{{Name: "infra/install.sh"}}); err != nil || ok {
		t.Fatalf("empty payload ok=%v err=%v", ok, err)
	}
}

func TestReplace(t *testing.T) {
	files := Replace([]File{{Name: "keep", Data: []byte("old")}, {Name: "replace", Data: []byte("old")}}, "replace", []byte("new"))
	if len(files) != 2 || files[0].Name != "keep" || files[1].Name != "replace" || string(files[1].Data) != "new" {
		t.Fatalf("files=%+v", files)
	}
}

func TestCollectExcludesDevelopmentArtifacts(t *testing.T) {
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
	files, err := Collect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "keep.txt" {
		t.Fatalf("files=%+v", files)
	}
}

func TestCollectRejectsSymlink(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "target"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", filepath.Join(root, "link")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := Collect(root); err == nil || !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("error=%v", err)
	}
}

func TestWriteCreatesParentsAndReplacesArtifact(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "app.zip")
	if err := Write(path, []byte("first")); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, []byte("second")); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "second" {
		t.Fatalf("content=%q", raw)
	}
}
