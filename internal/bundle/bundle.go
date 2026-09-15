package bundle

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type File struct {
	Name string
	Data []byte
}

var archiveEpoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func Collect(root string) ([]File, error) {
	var files []File
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		if entry.IsDir() {
			if relative == ".git" || relative == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not supported: %s", relative)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("special files are not supported: %s", relative)
		}
		if relative == ".gitignore" || relative == "infra/package.sh" || relative == "infra/payload.tar.gz" || strings.HasSuffix(relative, ".zip") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, File{Name: relative, Data: data})
		return nil
	})
	return files, err
}

func InfraPayload(files []File) ([]byte, bool, error) {
	var infra []File
	for _, file := range files {
		if strings.HasPrefix(file.Name, "infra/") && file.Name != "infra/install.sh" && file.Name != "infra/payload.tar.gz" {
			infra = append(infra, file)
		}
	}
	if len(infra) == 0 {
		return nil, false, nil
	}
	sort.Slice(infra, func(i, j int) bool { return infra[i].Name < infra[j].Name })
	var payload bytes.Buffer
	gzipWriter, _ := gzip.NewWriterLevel(&payload, gzip.BestCompression)
	gzipWriter.Header.ModTime = time.Unix(0, 0)
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	for _, file := range infra {
		header := &tar.Header{Name: file.Name, Mode: 0644, Size: int64(len(file.Data)), ModTime: time.Unix(0, 0), Uid: 0, Gid: 0, Format: tar.FormatUSTAR}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, false, err
		}
		if _, err := tarWriter.Write(file.Data); err != nil {
			return nil, false, err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, false, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, false, err
	}
	if payload.Len() > 8<<20 {
		return nil, false, fmt.Errorf("compressed infra payload exceeds 8 MiB")
	}
	return payload.Bytes(), true, nil
}

func Replace(files []File, name string, data []byte) []File {
	filtered := files[:0]
	for _, file := range files {
		if file.Name != name {
			filtered = append(filtered, file)
		}
	}
	return append(filtered, File{Name: name, Data: data})
}

func Archive(files []File) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for _, file := range files {
		header := &zip.FileHeader{Name: file.Name, Method: zip.Deflate}
		header.SetModTime(archiveEpoch)
		header.SetMode(0644)
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err := entry.Write(file.Data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return archive.Bytes(), nil
}

func Write(output string, archive []byte) error {
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(output), ".remote-build-*.zip")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err = temporary.Write(archive); err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, output)
}
