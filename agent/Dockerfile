# syntax=docker/dockerfile:1.4
FROM golang:alpine AS builder
LABEL authors="jaam8"

WORKDIR /build

COPY agent/go.* ./agent/
COPY common-lib/go.* ./common-lib/

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go work init ./common-lib ./agent && go mod download

COPY agent/ ./agent/
COPY common-lib/ ./common-lib/
COPY .env/ ./.env

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o agent-service agent/cmd/main.go

FROM gcr.io/distroless/base-debian12 AS runner

WORKDIR /app

COPY --from=builder /build/.env ../.env
COPY --from=builder /build/agent-service .

CMD ["./agent-service"]