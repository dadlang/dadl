package parser

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

//ResourceProvider provides additional reources
type ResourceProvider interface {
	GetResource(name string) (io.ReadCloser, error)
	FindResources(pattern string) ([]string, error)
	ForResource(name string) ResourceProvider
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

func (p fsResourceProvider) FindResources(pattern string) ([]string, error) {
	log.Print("find resources: ", pattern, " in ", p.basePath)
	res, err := filepath.Glob(filepath.Join(p.basePath, pattern))
	if err != nil {
		return nil, err
	}
	for i, r := range res {
		res[i], err = filepath.Rel(p.basePath, r)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (p fsResourceProvider) ForResource(relativePath string) ResourceProvider {
	log.Print("build provider for resource: ", relativePath)
	return NewFSResourceProvider(filepath.Join(p.basePath, filepath.Dir(relativePath)))
}
