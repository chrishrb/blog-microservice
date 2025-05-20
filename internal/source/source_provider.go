package source

import (
	"fmt"
	"io"
	"os"
)

type SourceProvider interface {
	GetData() ([]byte, error)
}

type StringSourceProvider struct {
	Data string
}

func (s StringSourceProvider) GetData() ([]byte, error) {
	return []byte(s.Data), nil
}

type FileSourceProvider struct {
	FileName string
}

func (f FileSourceProvider) GetData() ([]byte, error) {
	file, err := os.Open(f.FileName)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %v", f.FileName, err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %v", f.FileName, err)
	}
	return data, nil
}
