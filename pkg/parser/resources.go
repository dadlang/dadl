package parser

import (
	"io"
	"os"
	"path/filepath"
)

//ResourceProvider provides additional reources
type ResourceProvider interface {
	GetResource(name string) (io.ReadCloser, error)
}

//NewFSResourceProvider creates new resource provider with given base path
func NewFSResourceProvider(basePath string) ResourceProvider {
	return fsResourceProvider{basePath: basePath}
}

type fsResourceProvider struct {
	basePath string
}

func (p fsResourceProvider) GetResource(relativePath string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(p.basePath, relativePath))
}
