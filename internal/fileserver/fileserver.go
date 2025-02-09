package fileserver

import (
	"net/http"
)

type FileServer struct {
	Dir    string
	Prefix string
}

func (fs *FileServer) ServeHTTP() http.Handler {
	return http.StripPrefix(
		fs.Prefix,
		http.FileServer(http.Dir(fs.Dir)))
}
