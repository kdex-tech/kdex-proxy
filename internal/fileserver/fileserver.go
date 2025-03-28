// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
