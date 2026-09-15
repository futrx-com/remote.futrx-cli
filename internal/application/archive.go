package application

import (
	"archive/zip"
	"bytes"
	"sort"
	"time"
)

var archiveEpoch = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

func buildApplicationArchive(files []sourceFile) ([]byte, error) {
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)
	for _, file := range files {
		header := &zip.FileHeader{Name: file.name, Method: zip.Deflate}
		header.SetModTime(archiveEpoch)
		header.SetMode(0644)
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return nil, err
		}
		if _, err = writer.Write(file.data); err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return archive.Bytes(), nil
}
