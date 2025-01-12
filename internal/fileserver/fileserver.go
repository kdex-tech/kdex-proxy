package fileserver

import (
	"fmt"
	"net/http"
	"os"
)

const (
	DefaultPrefix = "/~/m/"
	DefaultDir    = "/modules"
)

type FileServer struct {
	Dir    string
	Prefix string
}

func NewFileServerFromEnv() (*FileServer, error) {
	dir := os.Getenv("MODULES_DIR")
	if dir == "" {
		dir = DefaultDir
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("modules directory %s does not exist", dir)
	}

	prefix := os.Getenv("MODULES_PREFIX")
	if prefix == "" {
		prefix = DefaultPrefix
	}

	return &FileServer{Dir: dir, Prefix: prefix}, nil
}

func (fs *FileServer) ServeHTTP() http.Handler {
	return http.StripPrefix(
		fs.Prefix,
		http.FileServer(http.Dir(fs.Dir)))
}
