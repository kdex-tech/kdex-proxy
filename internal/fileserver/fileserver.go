package fileserver

import (
	"net/http"

	"kdex.dev/proxy/internal/config"
)

type FileServer struct {
	Dir    string
	Prefix string
}

func NewFileServer(config *config.Config) *FileServer {
	return &FileServer{
		Dir:    config.ModuleDir,
		Prefix: config.Fileserver.Prefix,
	}
}

func (fs *FileServer) ServeHTTP() http.Handler {
	return http.StripPrefix(
		fs.Prefix,
		http.FileServer(http.Dir(fs.Dir)))
}
