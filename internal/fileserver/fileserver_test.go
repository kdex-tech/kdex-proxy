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
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileServer_ServeHTTP(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir = filepath.Join(dir, "..", "..", "test/resources")

	type fields struct {
		Dir    string
		Prefix string
	}
	tests := []struct {
		name   string
		fields fields
		path   string
		status int
		body   string
	}{
		{
			name:   "default",
			fields: fields{Dir: dir, Prefix: "/bar"},
			path:   "/bar/index.js",
			status: http.StatusOK,
			body:   `console.log("hello");`,
		},
		{
			name:   "file not found",
			fields: fields{Dir: dir, Prefix: "/bar"},
			path:   "/bar/foo.js",
			status: http.StatusNotFound,
			body:   "404 page not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &FileServer{
				Dir:    tt.fields.Dir,
				Prefix: tt.fields.Prefix,
			}
			handler := fs.ServeHTTP()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", tt.path, nil)
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, recorder.Code, tt.status)
			assert.Equal(t, recorder.Body.String(), tt.body)
		})
	}
}
