package application

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type sourceFile struct {
	name string
	data []byte
}

func collectSourceFiles(root string) ([]sourceFile, error) {
	var files []sourceFile
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, _ := filepath.Rel(root, path)
		relativePath = filepath.ToSlash(relativePath)
		if entry.IsDir() {
			if relativePath == ".git" || relativePath == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlinks are not supported: %s", relativePath)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("special files are not supported: %s", relativePath)
		}
		if relativePath == ".gitignore" || relativePath == "infra/package.sh" || relativePath == "infra/payload.tar.gz" || strings.HasSuffix(relativePath, ".zip") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, sourceFile{name: relativePath, data: data})
		return nil
	})
	return files, err
}

func withoutSourceFile(files []sourceFile, name string) []sourceFile {
	filtered := files[:0]
	for _, file := range files {
		if file.name != name {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
