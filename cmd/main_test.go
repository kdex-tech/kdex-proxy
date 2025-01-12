// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_middlewareLogger(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				}),
			},
			want: "Hello, World!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := middlewareLogger(tt.args.next)
			if got == nil {
				t.Errorf("middlewareLogger() = %v", got)
				return
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			got.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Errorf("middlewareLogger().Code = %v", recorder.Code)
			}
			if recorder.Body.String() != tt.want {
				t.Errorf("middlewareLogger().Body = %v", recorder.Body.String())
			}
		})
	}
}
