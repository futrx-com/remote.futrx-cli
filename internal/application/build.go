package application

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type BuildResult struct {
	Path, ID, Version, SHA256 string
	Files                     int
}
type sourceFile struct {
	name string
	data []byte
}

var epoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func Build(dir, output string) (BuildResult, error) {
	m, err := Validate(dir)
	if err != nil {
		return BuildResult{}, err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return BuildResult{}, err
	}
	if output == "" {
		output = filepath.Join(filepath.Dir(dir), fmt.Sprintf("%s-%s.zip", m.ID, safeVersion(m.Version)))
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return BuildResult{}, err
	}
	if strings.HasPrefix(output+string(os.PathSeparator), dir+string(os.PathSeparator)) {
		return BuildResult{}, fmt.Errorf("output must be outside the application directory")
	}
	files, err := collect(dir)
	if err != nil {
		return BuildResult{}, err
	}
	payload, hasInfra, err := infraPayload(files)
	if err != nil {
		return BuildResult{}, err
	}
	if hasInfra {
		files = without(files, "infra/payload.tar.gz")
		files = append(files, sourceFile{"infra/payload.tar.gz", payload})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	for _, f := range files {
		h := &zip.FileHeader{Name: f.name, Method: zip.Deflate}
		h.SetModTime(epoch)
		h.SetMode(0644)
		w, e := zw.CreateHeader(h)
		if e != nil {
			return BuildResult{}, e
		}
		if _, e = w.Write(f.data); e != nil {
			return BuildResult{}, e
		}
	}
	if err := zw.Close(); err != nil {
		return BuildResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return BuildResult{}, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(output), ".remote-build-*.zip")
	if err != nil {
		return BuildResult{}, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err = tmp.Write(archive.Bytes()); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return BuildResult{}, err
	}
	if err = os.Rename(tmpName, output); err != nil {
		return BuildResult{}, err
	}
	sum := sha256.Sum256(archive.Bytes())
	return BuildResult{output, m.ID, m.Version, hex.EncodeToString(sum[:]), len(files)}, nil
}

func collect(root string) ([]sourceFile, error) {
	var out []sourceFile
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == ".git" || rel == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not supported: %s", rel)
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("special files are not supported: %s", rel)
		}
		if rel == ".gitignore" || rel == "infra/package.sh" || rel == "infra/payload.tar.gz" || strings.HasSuffix(rel, ".zip") {
			return nil
		}
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		out = append(out, sourceFile{rel, b})
		return nil
	})
	return out, err
}

func infraPayload(files []sourceFile) ([]byte, bool, error) {
	var infra []sourceFile
	for _, f := range files {
		if strings.HasPrefix(f.name, "infra/") && f.name != "infra/install.sh" && f.name != "infra/payload.tar.gz" {
			infra = append(infra, f)
		}
	}
	if len(infra) == 0 {
		return nil, false, nil
	}
	sort.Slice(infra, func(i, j int) bool { return infra[i].name < infra[j].name })
	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gz.Header.ModTime = time.Unix(0, 0)
	gz.Header.OS = 255
	tw := tar.NewWriter(gz)
	for _, f := range infra {
		h := &tar.Header{Name: f.name, Mode: 0644, Size: int64(len(f.data)), ModTime: time.Unix(0, 0), Uid: 0, Gid: 0, Format: tar.FormatUSTAR}
		if err := tw.WriteHeader(h); err != nil {
			return nil, false, err
		}
		if _, err := tw.Write(f.data); err != nil {
			return nil, false, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, false, err
	}
	if err := gz.Close(); err != nil {
		return nil, false, err
	}
	if buf.Len() > 8<<20 {
		return nil, false, fmt.Errorf("compressed infra payload exceeds 8 MiB")
	}
	return buf.Bytes(), true, nil
}
func without(in []sourceFile, name string) []sourceFile {
	out := in[:0]
	for _, f := range in {
		if f.name != name {
			out = append(out, f)
		}
	}
	return out
}
func safeVersion(v string) string {
	var b strings.Builder
	for _, r := range v {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("._-", r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
