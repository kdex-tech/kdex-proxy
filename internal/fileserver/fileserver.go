package fileserver

import (
	"fmt"
	"log"
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
	dir := os.Getenv("MODULE_DIR")
	if dir == "" {
		dir = DefaultDir
		log.Printf("Defaulting module_dir to %s", dir)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("module directory %s does not exist", dir)
	}

	prefix := os.Getenv("MODULE_PREFIX")
	if prefix == "" {
		prefix = DefaultPrefix
		log.Printf("Defaulting module_prefix to %s", prefix)
	}

	return &FileServer{Dir: dir, Prefix: prefix}, nil
}

func (fs *FileServer) ServeHTTP() http.Handler {
	return http.StripPrefix(
		fs.Prefix,
		http.FileServer(http.Dir(fs.Dir)))
}
