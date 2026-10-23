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
# species.json; stage `test` runs the tests against it; stage `builder` compiles
# the server binary once they pass; the final image
# copies only the static binary and the compiled dataset (tech.twoCommandsOneModule, tech.selfContainedServer, tech.datasetNeverCommitted).
# Each stage copies every package its command imports (tech.containerImageBuilds);
# TestDockerStagesCopyWhatTheyBuild checks that, so a new import needs a COPY here.
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

# Stage `test` runs every test but the spec checks against the dataset the
# image will carry, and the server build below depends on it, so a failing
# test stops the deploy (tech.testedBeforeDeploy). The spec checks read git
# history, which a build context lacks; they run on every push and pull
# request instead (tech.checksOnEveryChange). Tests that need Chrome skip here.
#
# To deploy despite failing tests, put an empty SKIP_TESTS file in the build
# context (never commit it). The build log says so.
FROM golang:1.26.9 AS test

WORKDIR /src
COPY . .
COPY --from=ingest /out/species.json web/data/species.json

RUN if [ -e SKIP_TESTS ]; then \
      echo '!!! SKIP_TESTS is present: this image was built WITHOUT running the tests (tech.testedBeforeDeploy)'; \
    else \
      go test $(go list ./... | grep -v '/river/check$'); \
    fi && \
    touch /tested

FROM golang:1.26.9 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/birdview ./cmd/birdview
COPY internal ./internal
COPY data ./data
COPY web ./web
COPY docs ./docs
# Depend on the test stage, so the server is not built until it has passed.
COPY --from=test /tested /tested

RUN CGO_ENABLED=0 go build -o /usr/local/bin/birdview ./cmd/birdview

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/local/bin/birdview /usr/local/bin/birdview
COPY --from=ingest  /out/species.json       /data/species.json

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/birdview"]
CMD ["-addr", "0.0.0.0:8080", "-data", "/data/species.json"]
