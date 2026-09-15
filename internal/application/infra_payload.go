package application

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"sort"
	"strings"
	"time"
)

func infraPayload(files []sourceFile) ([]byte, bool, error) {
	var infra []sourceFile
	for _, file := range files {
		if strings.HasPrefix(file.name, "infra/") && file.name != "infra/install.sh" && file.name != "infra/payload.tar.gz" {
			infra = append(infra, file)
		}
	}
	if len(infra) == 0 {
		return nil, false, nil
	}
	sort.Slice(infra, func(i, j int) bool { return infra[i].name < infra[j].name })

	var payload bytes.Buffer
	gzipWriter, _ := gzip.NewWriterLevel(&payload, gzip.BestCompression)
	gzipWriter.Header.ModTime = time.Unix(0, 0)
	gzipWriter.Header.OS = 255
	tarWriter := tar.NewWriter(gzipWriter)
	for _, file := range infra {
		header := &tar.Header{Name: file.name, Mode: 0644, Size: int64(len(file.data)), ModTime: time.Unix(0, 0), Uid: 0, Gid: 0, Format: tar.FormatUSTAR}
		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, false, err
		}
		if _, err := tarWriter.Write(file.data); err != nil {
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
