// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package util

import (
	"bytes"
	"math"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"golang.org/x/net/html"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func Filter[T any](slice []T, filter func(T) bool) []T {
	filtered := make([]T, 0, len(slice))
	for _, item := range slice {
		if filter(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func GetScheme(r *http.Request) string {
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	return "http"
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func NormalizeString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "")
	s = strings.TrimSpace(s)
	return s
}

func RandStringBytes(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TimeFromFloat64Seconds(seconds float64) time.Time {
	round, frac := math.Modf(seconds)
	return time.Unix(int64(round), int64(frac*1e9)).Truncate(time.Second)
}

func ToDoc(body string) *html.Node {
	doc, _ := html.Parse(bytes.NewReader([]byte(body)))
	return doc
}
