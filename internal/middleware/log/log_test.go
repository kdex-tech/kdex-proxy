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

package log

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_statusRecorder_WriteHeader(t *testing.T) {
	type fields struct {
		ResponseWriter http.ResponseWriter
		status         int
	}
	type args struct {
		code int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "status 200",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				status:         200,
			},
			args: args{
				code: http.StatusOK,
			},
		},
		{
			name: "status 302",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				status:         302,
			},
			args: args{
				code: http.StatusFound,
			},
		},
		{
			name: "status 404",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				status:         404,
			},
			args: args{
				code: http.StatusNotFound,
			},
		},
		{
			name: "status 500",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				status:         500,
			},
			args: args{
				code: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &statusRecorder{
				ResponseWriter: tt.fields.ResponseWriter,
				status:         tt.fields.status,
			}
			rec.WriteHeader(tt.args.code)
			assert.Equal(t, tt.fields.status, rec.status)
		})
	}
}

type logRecorder struct {
	log.Logger
	wLogger     *log.Logger
	capturedLog string
}

func (l *logRecorder) Printf(format string, v ...any) {
	l.capturedLog = fmt.Sprintf(format, v...)
	l.wLogger.Printf(format, v...)
}

func Test_Log(t *testing.T) {
	type args struct {
		path   string
		method string
		next   http.Handler
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "GET returns 200",
			args: args{
				path:   "/",
				method: "GET",
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
					w.WriteHeader(http.StatusOK)
				}),
			},
			want: "GET / status 200, processed in ",
		},
		{
			name: "POST foo returns 404",
			args: args{
				path:   "/foo",
				method: "POST",
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
			},
			want: "POST /foo status 404, processed in ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logRecorder := &logRecorder{
				wLogger: log.Default(),
			}
			loggerMW := &LoggerMiddleware{
				Impl: logRecorder,
			}

			got := loggerMW.Log(tt.args.next, false)
			if got == nil {
				t.Errorf("Log() = %v", got)
				return
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			got.ServeHTTP(recorder, request)
			assert.Contains(t, logRecorder.capturedLog, tt.want)
		})
	}
}
