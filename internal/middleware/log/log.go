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
	"net/http"
	"time"
)

type LoggerMiddleware struct {
	Impl interface {
		Printf(format string, v ...any)
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (l *LoggerMiddleware) Log(next http.Handler, onlyFailures bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := statusRecorder{
			ResponseWriter: w,
			status:         200,
		}

		start := time.Now()
		defer func() {
			if onlyFailures && wrapper.status < 400 {
				return
			}
			l.Impl.Printf("%s %s status %d, processed in %v", r.Method, r.URL.Path, wrapper.status, time.Since(start))
		}()

		next.ServeHTTP(&wrapper, r)
	})
}
