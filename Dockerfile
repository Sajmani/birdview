# syntax=docker/dockerfile:1
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Multi-stage build: stage `ingest` compiles and runs birdview-ingest to produce
# species.json; stage `builder` compiles the server binary; the final image
# copies only the static binary and the compiled dataset (T-001, T-002, T-005).
#
# Usage:
#   docker build -t birdview .
#   docker run --rm -p 8080:8080 birdview

FROM golang:1.26.9 AS ingest

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/birdview-ingest ./cmd/birdview-ingest
COPY internal ./internal
COPY data ./data
COPY third_party ./third_party

RUN CGO_ENABLED=0 go build -o /usr/local/bin/birdview-ingest ./cmd/birdview-ingest && \
    mkdir -p /out && \
    birdview-ingest -out /out/species.json

FROM golang:1.26.9 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/birdview ./cmd/birdview
COPY internal ./internal
COPY web ./web
COPY spec ./spec
COPY docs ./docs

RUN CGO_ENABLED=0 go build -o /usr/local/bin/birdview ./cmd/birdview

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/local/bin/birdview /usr/local/bin/birdview
COPY --from=ingest  /out/species.json       /data/species.json

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/birdview"]
CMD ["-addr", "0.0.0.0:8080", "-data", "/data/species.json"]
