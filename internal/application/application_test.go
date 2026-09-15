package application

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldValidateAndBuildReproducibly(t *testing.T) {
	parent := t.TempDir()
	dir, err := Scaffold(parent, "Demo App")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Validate(dir)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ID != "demo-app" {
		t.Fatalf("id = %q", manifest.ID)
	}
	if err := os.WriteFile(filepath.Join(dir, "infra", "package.sh"), []byte("#!/bin/sh\nexit 99\n"), 0755); err != nil {
		t.Fatal(err)
	}

	one := filepath.Join(parent, "one.zip")
	two := filepath.Join(parent, "two.zip")
	first, err := Build(dir, one)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(dir, two)
	if err != nil {
		t.Fatal(err)
	}
	if first.SHA256 != second.SHA256 {
		t.Fatalf("build is not reproducible: %s != %s", first.SHA256, second.SHA256)
	}

	raw, err := os.ReadFile(one)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"application.json": false, "infra/install.sh": false, "infra/payload.tar.gz": false, "backend/main.go": false, "ui/scripts/main.js": false, "skills/demo-app/SKILL.md": false}
	for _, file := range zr.File {
		if _, ok := want[file.Name]; ok {
			want[file.Name] = true
		}
		if file.Name == ".gitignore" {
			t.Error("development .gitignore was packaged")
		}
		if file.Name == "infra/package.sh" {
			t.Error("legacy package.sh was packaged")
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("archive missing %s", name)
		}
	}
}

func TestValidateRejectsUnsafeInstallPath(t *testing.T) {
	dir := t.TempDir()
	raw := `{"id":"bad","name":"Bad","version":"1","scopes":["project"],"install":"../install.sh"}`
	if err := os.WriteFile(filepath.Join(dir, "application.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(dir); err == nil {
		t.Fatal("Validate accepted install path outside infra")
	}
}
