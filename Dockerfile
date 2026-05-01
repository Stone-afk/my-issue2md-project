# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.24.7-bookworm AS builder
WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH

COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY cmd/issue2md ./cmd/issue2md
COPY cmd/issue2mdweb ./cmd/issue2mdweb
COPY internal ./internal

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/bin/issue2md ./cmd/issue2md

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/bin/issue2mdweb ./cmd/issue2mdweb

RUN chmod +x /out/bin/issue2md /out/bin/issue2mdweb

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app -h /work app && apk add --no-cache ca-certificates

WORKDIR /work

COPY --from=builder /out/bin/issue2md /app/bin/issue2md
COPY --from=builder /out/bin/issue2mdweb /app/bin/issue2mdweb
COPY docker-entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh /app/bin/issue2md /app/bin/issue2mdweb \
    && mkdir -p /work /app/bin \
    && chown -R app:app /app /work

USER app:app
ENTRYPOINT ["/app/entrypoint.sh"]
