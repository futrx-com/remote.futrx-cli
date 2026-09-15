package application

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BuildResult struct {
	Path    string
	ID      string
	Version string
	SHA256  string
	Files   int
}

func Build(dir, output string) (BuildResult, error) {
	manifest, err := Validate(dir)
	if err != nil {
		return BuildResult{}, err
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return BuildResult{}, err
	}
	if output == "" {
		output = filepath.Join(filepath.Dir(dir), fmt.Sprintf("%s-%s.zip", manifest.ID, safeVersion(manifest.Version)))
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return BuildResult{}, err
	}
	if strings.HasPrefix(output+string(os.PathSeparator), dir+string(os.PathSeparator)) {
		return BuildResult{}, fmt.Errorf("output must be outside the application directory")
	}
	files, err := collectSourceFiles(dir)
	if err != nil {
		return BuildResult{}, err
	}
	payload, hasInfra, err := infraPayload(files)
	if err != nil {
		return BuildResult{}, err
	}
	if hasInfra {
		files = withoutSourceFile(files, "infra/payload.tar.gz")
		files = append(files, sourceFile{name: "infra/payload.tar.gz", data: payload})
	}
	archive, err := buildApplicationArchive(files)
	if err != nil {
		return BuildResult{}, err
	}
	if err := writeBuildArtifact(output, archive); err != nil {
		return BuildResult{}, err
	}
	sum := sha256.Sum256(archive)
	return BuildResult{
		Path:    output,
		ID:      manifest.ID,
		Version: manifest.Version,
		SHA256:  hex.EncodeToString(sum[:]),
		Files:   len(files),
	}, nil
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
