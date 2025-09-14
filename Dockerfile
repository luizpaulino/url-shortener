# ---------- builder ----------
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Cache de deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
ARG VERSION=0.1.0
ARG COMMIT=dev
ARG BUILD_DATE=unknown
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'" \
    -o /out/shortener-api ./cmd/api

# ---------- final ----------
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /out/shortener-api /shortener-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/shortener-api"]
