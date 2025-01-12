package fileserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileServerFromEnv(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	dir = filepath.Join(dir, "..", "..", "test/resources")

	type args struct {
		dir    string
		prefix string
	}

	tests := []struct {
		name   string
		args   args
		dir    string
		prefix string
		err    error
	}{
		{
			name:   "default",
			args:   args{dir: dir, prefix: "/bar/"},
			dir:    dir,
			prefix: "/bar/",
			err:    nil,
		},
		{
			name:   "dir does not exist",
			args:   args{dir: "/foo", prefix: "/bar/"},
			dir:    "/foo",
			prefix: "/bar/",
			err:    fmt.Errorf("module directory %s does not exist", "/foo"),
		},
		{
			name:   "prefix is empty",
			args:   args{dir: dir},
			dir:    dir,
			prefix: DefaultPrefix,
			err:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MODULE_DIR", tt.args.dir)
			os.Setenv("MODULE_PREFIX", tt.args.prefix)
			got, err := NewFileServerFromEnv()
			assert.Equal(t, err, tt.err)
			if tt.err != nil {
				return
			}
			assert.Equal(t, got.Dir, tt.dir)
			assert.Equal(t, got.Prefix, tt.prefix)
		})
	}
}

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
