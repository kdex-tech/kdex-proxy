# Copyright 2025 KDex Tech
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build the manager binary
ARG GO_VERSION=1.22

FROM golang:${GO_VERSION} AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Install delve debugger
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -gcflags "all=-N -l" -a -o proxy cmd/main.go

# Final stage
FROM gcr.io/distroless/static:nonroot

LABEL org.opencontainers.image.source=https://github.com/kdex-tech/kdex-proxy
LABEL org.opencontainers.image.description="KDex Proxy"
LABEL org.opencontainers.image.licenses=Apache-2.0

COPY --from=builder /go/bin/dlv /dlv
COPY --from=builder /proxy /proxy

# Expose port
EXPOSE 8080

USER 65532:65532

# Run the binary
ENTRYPOINT ["/dlv"]
