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

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/app"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/importmap"
	"kdex.dev/proxy/internal/meta"
	"kdex.dev/proxy/internal/navigation"
	"kdex.dev/proxy/internal/store/cache"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

type Proxy struct {
	Config      *config.Config
	cache       *cache.CacheStore
	transformer transform.Transformer
}

func NewProxy(config *config.Config) *Proxy {
	transformer := &transform.AggregatedTransformer{
		Transformers: []transform.Transformer{
			importmap.NewImportMapTransformer(config),
			meta.NewMetaTransformer(config),
			navigation.NewNavigationTransformer(config),
			app.NewAppTransformer(config),
		},
	}

	cache := cache.NewCacheStore(config)

	return &Proxy{
		Config:      config,
		cache:       cache,
		transformer: transformer,
	}
}

func (s *Proxy) Probe(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s://%s%s", util.GetScheme(r), s.Config.Proxy.UpstreamAddress, s.Config.Proxy.UpstreamHealthzPath)

	req, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		w.Write([]byte("OK"))
	} else {
		w.Write([]byte(fmt.Sprintf("GET %s returned %d", s.Config.Proxy.UpstreamHealthzPath, resp.StatusCode)))
	}
}

func (s *Proxy) ReverseProxy() func(http.ResponseWriter, *http.Request) {
	rp := &httputil.ReverseProxy{
		ErrorHandler:   s.errorHandler,
		ModifyResponse: s.modifyResponse,
		Rewrite:        s.rewrite,
	}

	return rp.ServeHTTP
}

func (s *Proxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("Error: %v", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Proxy) joinURLPath(a, b *url.URL) (path, rawPath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return s.singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	aPath := a.EscapedPath()
	bPath := b.EscapedPath()

	aSlash := strings.HasSuffix(aPath, "/")
	bSlash := strings.HasPrefix(bPath, "/")

	switch {
	case aSlash && bSlash:
		return a.Path + b.Path[1:], aPath + bPath[1:]
	case !aSlash && !bSlash:
		return a.Path + "/" + b.Path, aPath + "/" + bPath
	}
	return a.Path + b.Path, aPath + bPath
}

func (s *Proxy) modifyResponse(r *http.Response) error {
	proxiedEtag, ok := r.Request.Context().Value(kctx.ProxiedEtagKey).(string)

	if !ok {
		proxiedEtag = ""
	}

	configHash := s.Config.Hash()
	cacheHit := false

	if r.StatusCode == http.StatusNotFound {
		return nil
	}

	// For 304 responses, we only need to handle ETag
	if r.StatusCode == http.StatusNotModified {
		upstreamETag := r.Header.Get("ETag")
		if upstreamETag != "" {
			derivedETag := fmt.Sprintf(`%s-t%x`, upstreamETag, configHash)

			if derivedETag == proxiedEtag {
				r.Header.Set("Cache-Control", "no-cache")
				r.Header.Set("ETag", derivedETag)
				r.Header.Add("Vary", "Authorization")

				return nil
			}

			// Check if we have cached content
			if s.cache != nil {
				if entry, _ := (*s.cache).Get(r.Request.Context(), upstreamETag); entry != nil {
					cacheHit = true
					// Config changed, need to transform cached content
					r.StatusCode = entry.StatusCode
					r.Body = io.NopCloser(bytes.NewReader(entry.Content))
					r.ContentLength = int64(len(entry.Content))
					r.Header.Set("Content-Type", entry.ContentType)
					r.Header.Set("Content-Length", strconv.Itoa(len(entry.Content)))
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	// For other responses, check if it's HTML we need to transform
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	upstreamETag := r.Header.Get("ETag")
	if upstreamETag != "" {
		derivedETag := fmt.Sprintf(`%s-t%x`, upstreamETag, configHash)
		r.Header.Set("ETag", derivedETag)
	}

	// Check for chunked transfer encoding
	isChunked := len(r.TransferEncoding) > 0 && r.TransferEncoding[0] == "chunked"

	// Transform content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	r.Body.Close()

	// Cache the original content
	if upstreamETag != "" && !cacheHit {
		if s.cache != nil {
			(*s.cache).Set(r.Request.Context(), upstreamETag, cache.CacheEntry{
				Content:     body,
				ContentType: r.Header.Get("Content-Type"),
				CreatedAt:   time.Now(),
				ETag:        upstreamETag,
				StatusCode:  r.StatusCode,
			})
		}
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	if err := s.transformer.Transform(r, doc); err != nil {
		return fmt.Errorf("failed to transform response: %w", err)
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fmt.Errorf("failed to render HTML: %w", err)
	}

	transformedBody := buf.Bytes()
	r.Body = io.NopCloser(bytes.NewReader(transformedBody))

	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Add("Vary", "Authorization")

	// Handle transfer encoding
	if isChunked {
		r.TransferEncoding = []string{"chunked"}
		r.Header.Del("Content-Length")
	} else {
		r.ContentLength = int64(len(transformedBody))
		r.Header.Set("Content-Length", strconv.Itoa(len(transformedBody)))
	}

	return nil
}

func (s *Proxy) rewrite(r *httputil.ProxyRequest) {
	target := &url.URL{
		Scheme:   s.Config.Proxy.UpstreamScheme,
		Host:     s.Config.Proxy.UpstreamAddress,
		Path:     s.Config.Proxy.UpstreamPrefix,
		RawQuery: r.In.URL.RawQuery,
	}

	req := r.Out

	proxiedParts := s.rewritePath(r)

	targetQuery := target.RawQuery
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host

	if proxiedParts.AppAlias != "" || proxiedParts.ProxiedPath != req.URL.Path {
		req.URL.Path = proxiedParts.ProxiedPath
	} else {
		req.URL.Path = r.In.URL.Path
	}

	// Once the path is rewritten we need to check if there are Config.Navigation.TemplatePaths that match the path
	// If there is a match we need to set the AppAlias and AppPath

	strippedURLPath := strings.TrimSuffix(req.URL.Path, "/")

	for _, templatePath := range s.Config.Navigation.TemplatePaths {
		if strings.HasPrefix(strippedURLPath, templatePath.Href) {
			req.URL.Path = templatePath.Template + strings.TrimPrefix(req.URL.Path, templatePath.Href)
		}
	}

	req.URL.Path, req.URL.RawPath = s.joinURLPath(target, req.URL)

	if strings.HasSuffix(req.URL.Path, "/") &&
		s.Config.Proxy.AppendIndex &&
		s.Config.Proxy.IndexFile != "" {

		req.URL.Path = req.URL.Path + s.Config.Proxy.IndexFile
	}

	log.Printf("Path rewritten '%s' to '%s', Alias: '%s', Path: '%s', Upstream: '%s'", r.In.URL.Path, req.URL.Path, proxiedParts.AppAlias, proxiedParts.AppPath, req.URL.String())

	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}

	req.Host = target.Host

	// Strip transformer suffix from If-None-Match header
	if ifNoneMatch := r.In.Header.Get("If-None-Match"); ifNoneMatch != "" {
		if idx := strings.LastIndex(ifNoneMatch, "-t"); idx != -1 {
			originalETag := ifNoneMatch[:idx]
			req.Header.Set("If-None-Match", originalETag)
			req = req.WithContext(context.WithValue(req.Context(), kctx.ProxiedEtagKey, ifNoneMatch))
		}
	}

	req = req.WithContext(context.WithValue(req.Context(), kctx.ProxiedPartsKey, proxiedParts))

	{
		// Everything bellow is about Proxy Protocol
		req.Header["X-Forwarded-For"] = r.In.Header["X-Forwarded-For"]
		r.SetXForwarded()
		setXForwardedPort(r.In, req)
		setForwarded(r.In, req)
	}

	r.Out = req
}

func (s *Proxy) proxiedPath(accept string, proxiedPath string) string {
	if s.Config.Proxy.AlwaysAppendSlash && (strings.Contains(accept, "text/html") ||
		strings.Contains(accept, "application/xhtml+xml") ||
		strings.Contains(accept, "application/xml")) &&
		!strings.HasSuffix(proxiedPath, ".html") &&
		!strings.HasSuffix(proxiedPath, ".htm") &&
		!strings.HasSuffix(proxiedPath, ".xhtml") &&
		!strings.HasSuffix(proxiedPath, ".xml") &&
		!strings.HasSuffix(proxiedPath, "/") {

		proxiedPath = proxiedPath + "/"
	}

	return proxiedPath
}

func (s *Proxy) rewritePath(r *httputil.ProxyRequest) kctx.ProxiedParts {
	accept := r.In.Header.Get("Accept")
	parts := strings.SplitN(r.In.URL.Path, s.Config.Proxy.PathSeparator, 2)

	if len(parts) > 1 {
		proxiedPath := s.proxiedPath(accept, parts[0])

		appParts := strings.SplitN(parts[1], "/", 2)

		if len(appParts) > 1 {
			appAlias := appParts[0]
			appPath := fmt.Sprintf("/%s", appParts[1])
			return kctx.ProxiedParts{
				AppAlias:    appAlias,
				AppPath:     appPath,
				ProxiedPath: proxiedPath,
			}
		} else {
			return kctx.ProxiedParts{
				AppAlias:    parts[1],
				AppPath:     "",
				ProxiedPath: proxiedPath,
			}
		}
	}

	return kctx.ProxiedParts{
		AppAlias:    "",
		AppPath:     "",
		ProxiedPath: s.proxiedPath(accept, r.In.URL.Path),
	}
}

func (s *Proxy) singleJoiningSlash(a, b string) string {
	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")
	switch {
	case aSlash && bSlash:
		return a + b[1:]
	case !aSlash && !bSlash:
		return a + "/" + b
	}
	return a + b
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IPv4zero
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func setForwarded(in *http.Request, out *http.Request) {
	out.Header.Set(
		"Forwarded",
		fmt.Sprintf(
			"by=%s;for=%s;host=%s;proto=%s",
			getOutboundIP().String(),
			out.Header.Get("X-Forwarded-For"),
			in.Host,
			util.GetScheme(in)))
}

func setXForwardedPort(in *http.Request, out *http.Request) {
	if strings.Contains(in.Host, ":") {
		_, port, err := net.SplitHostPort(in.Host)
		if err != nil {
			log.Printf("Error splitting host and port from %s: %v", in.Host, err)
			return
		}
		out.Header.Set("X-Forwarded-Port", port)
	}
}
